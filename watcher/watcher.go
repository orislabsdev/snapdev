// Package watcher monitors a directory tree for file changes and emits
// debounced ChangeEvents on a channel consumed by the build pipeline.
package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/orislabsdev/snapdev/config"
	"github.com/orislabsdev/snapdev/logger"
)

// ChangeEvent describes a single file-system change that passed all filters.
type ChangeEvent struct {
	// Path is the absolute path of the changed file.
	Path string
	// Operation is the human-readable fsnotify operation (e.g. "WRITE", "CREATE").
	Operation string
}

// Watcher wraps fsnotify and adds:
//   - recursive directory watching
//   - extension-based filtering
//   - ignore-pattern filtering
//   - debouncing to collapse rapid saves into a single rebuild trigger
type Watcher struct {
	cfg    *config.Config
	log    *logger.Logger
	fsw    *fsnotify.Watcher
	Events chan ChangeEvent // Consumers read rebuild triggers from this channel.

	mu      sync.Mutex
	timer   *time.Timer
	stopped bool
}

// New creates and initialises a Watcher. Call Start to begin receiving events.
func New(cfg *config.Config, log *logger.Logger) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		cfg:    cfg,
		log:    log,
		fsw:    fsw,
		Events: make(chan ChangeEvent, 1),
	}, nil
}

// Start walks cfg.WatchDir recursively, registers each subdirectory with
// fsnotify, and launches the background event-processing goroutine.
//
// If the watch directory does not exist, Start returns an error immediately so
// the caller can surface a clear diagnostic message.
func (w *Watcher) Start() error {
	if _, err := os.Stat(w.cfg.WatchDir); os.IsNotExist(err) {
		return &ErrDirNotFound{Dir: w.cfg.WatchDir}
	}

	count := 0
	err := filepath.Walk(w.cfg.WatchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			w.log.Warn("Skipping %s: %v", path, err)
			return nil // Non-fatal; keep walking.
		}

		if info.IsDir() {
			if w.isIgnored(path) {
				w.log.Debug("Ignoring directory: %s", path)
				return filepath.SkipDir
			}
			if err := w.fsw.Add(path); err != nil {
				w.log.Warn("Could not watch %s: %v", path, err)
				return nil
			}
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}

	w.log.Info("Watching %d director(ies) under %q", count, w.cfg.WatchDir)
	go w.loop()
	return nil
}

// Stop shuts down the watcher and closes the Events channel.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped {
		return
	}
	w.stopped = true

	if w.timer != nil {
		w.timer.Stop()
	}
	_ = w.fsw.Close()
	close(w.Events)
}

// loop is the background goroutine that translates raw fsnotify events into
// debounced ChangeEvents sent on w.Events.
func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return // fsnotify channel closed; watcher is stopping.
			}
			w.handleFSEvent(event)

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			w.log.Error("fsnotify error: %v", err)
		}
	}
}

// handleFSEvent filters and debounces a raw fsnotify event.
func (w *Watcher) handleFSEvent(event fsnotify.Event) {
	// Skip CHMOD-only events; they don't affect source content.
	if event.Op == fsnotify.Chmod {
		return
	}

	// Only react to extensions we care about.
	if !w.isWatchedExtension(event.Name) {
		return
	}

	// Skip paths matching any ignore pattern.
	if w.isIgnored(event.Name) {
		return
	}

	w.log.Debug("FS event: %s %s", event.Op, event.Name)

	// Debounce: reset the timer on every qualifying event.
	// The ChangeEvent is only emitted after cfg.Debounce of silence.
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped {
		return
	}

	if w.timer != nil {
		w.timer.Stop()
	}

	ce := ChangeEvent{
		Path:      event.Name,
		Operation: event.Op.String(),
	}

	w.timer = time.AfterFunc(w.cfg.Debounce, func() {
		w.mu.Lock()
		stopped := w.stopped
		w.mu.Unlock()

		if stopped {
			return
		}

		// Non-blocking send: if the consumer hasn't processed the previous
		// event yet, we drop the older one and keep only the newest trigger.
		select {
		case w.Events <- ce:
		default:
			// Drain the stale event and enqueue the fresh one.
			select {
			case <-w.Events:
			default:
			}
			w.Events <- ce
		}
	})
}

// isIgnored returns true if path contains any of the configured ignore patterns.
func (w *Watcher) isIgnored(path string) bool {
	// Normalise to forward slashes for cross-platform consistency.
	normalised := filepath.ToSlash(path)
	for _, pattern := range w.cfg.Ignore {
		if strings.Contains(normalised, pattern) {
			return true
		}
	}
	return false
}

// isWatchedExtension returns true if path has an extension listed in cfg.Extensions.
// Directories (no extension) always return true so we can watch new sub-dirs.
func (w *Watcher) isWatchedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return true
	}
	for _, e := range w.cfg.Extensions {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

// ErrDirNotFound is returned when the configured watch directory is absent.
type ErrDirNotFound struct {
	Dir string
}

func (e *ErrDirNotFound) Error() string {
	return "watch directory not found: " + e.Dir
}