// Package server provides a lightweight HTTP server that:
//
//  1. Serves compiled static assets from cfg.OutputDir.
//  2. Implements SPA (Single-Page Application) fallback routing by returning
//     index.html for any path that does not resolve to a real file.
//  3. Optionally injects a small Server-Sent Events (SSE) client snippet into
//     every HTML response so that the browser reloads automatically after each
//     successful build.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/orislabsdev/snapdev/config"
	"github.com/orislabsdev/snapdev/logger"
)

// liveReloadSnippet is injected just before </body> in every HTML response
// when live reload is enabled. It connects to the SSE endpoint and calls
// location.reload() whenever the server emits a "reload" event.
//
// The snippet uses EventSource (SSE) rather than WebSockets so no external
// library is required and browser compatibility is excellent.
const liveReloadSnippet = `
<!-- snapdev live reload -->
<script>
(function () {
  var es = new EventSource('/__snapdev_sse');
  es.onmessage = function (e) {
    if (e.data === 'reload') {
      console.log('[snapdev] Reloading…');
      location.reload();
    }
  };
  es.onerror = function () {
    // The dev server restarted; wait a moment then reload to reconnect.
    setTimeout(function () { location.reload(); }, 2000);
  };
  console.log('[snapdev] Live reload connected');
})();
</script>
`

// Server manages the HTTP server, SSE client registry, and graceful shutdown.
type Server struct {
	cfg     *config.Config
	log     *logger.Logger
	httpSrv *http.Server

	// mu guards the sseClients map.
	mu         sync.Mutex
	sseClients map[chan string]struct{}
	proxy      *httputil.ReverseProxy
}

// New creates a Server. Call Start to begin accepting connections.
func New(cfg *config.Config, log *logger.Logger) *Server {
	s := &Server{
		cfg:        cfg,
		log:        log,
		sseClients: make(map[chan string]struct{}),
	}

	mux := http.NewServeMux()

	// Register routes.
	if cfg.LiveReload {
		mux.HandleFunc("/__snapdev_sse", s.handleSSE)
	}

	if cfg.ReverseProxy != "" {
		target, err := url.Parse(cfg.ReverseProxy)
		if err != nil {
			log.Warn("Invalid reverse proxy URL %q: %v (proxy disabled)", cfg.ReverseProxy, err)
		} else {
			// Initialize the proxy for the target host.
			s.proxy = httputil.NewSingleHostReverseProxy(target)

			// Fix: Many backend servers (like Vite/Next.js/Express) reject requests if the
			// Host header doesn't match their address, returning "400 Bad Request".
			// We override the Director to ensure the Host header is correctly set.
			originalDirector := s.proxy.Director
			s.proxy.Director = func(req *http.Request) {
				originalDirector(req)
				req.Host = target.Host
			}
		}
	}

	mux.HandleFunc("/", s.handleStatic)

	s.httpSrv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 0, // 0 = no timeout (required for SSE streaming).
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start begins serving HTTP on the configured address. It blocks until the
// server is shut down; use Shutdown to stop it gracefully.
func (s *Server) Start() error {
	url := fmt.Sprintf("http://%s:%d", s.cfg.Host, s.cfg.Port)
	s.log.Info("Serving %s → %s", s.cfg.OutputDir, url)

	ln, err := net.Listen("tcp", s.httpSrv.Addr)
	if err != nil {
		return fmt.Errorf("could not bind to %s: %w", s.httpSrv.Addr, err)
	}

	return s.httpSrv.Serve(ln)
}

// Shutdown gracefully stops the HTTP server, giving in-flight requests up to
// five seconds to complete.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.httpSrv.Shutdown(ctx)
}

// NotifyReload broadcasts a "reload" SSE event to every connected browser tab.
// It is called by the main event loop after a successful build.
func (s *Server) NotifyReload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := len(s.sseClients)
	for ch := range s.sseClients {
		// Non-blocking send to avoid stalling the build loop if a client is slow.
		select {
		case ch <- "reload":
		default:
		}
	}

	if count > 0 {
		s.log.Info("Triggered live reload on %d connected client(s)", count)
	}
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

// handleStatic resolves the request path to a file inside cfg.OutputDir and
// serves it. Unknown paths fall back to index.html for SPA client-side routing.
// HTML files have the live-reload snippet injected when live reload is enabled.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// 1. Sanitise the URL path and map it to a filesystem path.
	rel := filepath.Clean(r.URL.Path)
	target := filepath.Join(s.cfg.OutputDir, rel)

	// 2. Check if the path exists locally.
	info, err := os.Stat(target)

	// 3. If the path is a directory (like "/"), try to locate an index.html within it.
	// This ensures we serve the static web app instead of proxying the root path.
	if err == nil && info.IsDir() {
		indexPath := filepath.Join(target, "index.html")
		if _, errIdx := os.Stat(indexPath); errIdx == nil {
			target = indexPath
			info, _ = os.Stat(target)
			err = nil // Path now points to a valid file.
		}
	}

	// 4. If the file exists locally and is not a directory, serve it.
	if err == nil && !info.IsDir() {
		// Inject live-reload snippet into HTML responses.
		if s.cfg.LiveReload && strings.EqualFold(filepath.Ext(target), ".html") {
			s.serveHTML(w, r, target)
			return
		}
		http.ServeFile(w, r, target)
		return
	}

	// 5. If not found locally, forward the request to the reverse proxy if configured.
	// This is where API requests (e.g. /api/data) typically get handled.
	if s.proxy != nil {
		s.proxy.ServeHTTP(w, r)
		return
	}

	// 6. SPA Fallback: If no file exists and proxying didn't handle it, serve the
	// main index.html. This allows client-side routers (React Router, etc.) to work.
	target = filepath.Join(s.cfg.OutputDir, "index.html")
	if s.cfg.LiveReload {
		s.serveHTML(w, r, target)
		return
	}
	http.ServeFile(w, r, target)
}

// serveHTML reads target, injects the live-reload snippet before </body>, and
// writes the resulting HTML. If the file is missing, a 404 is returned.
func (s *Server) serveHTML(w http.ResponseWriter, r *http.Request, target string) {
	raw, err := os.ReadFile(target)
	if err != nil {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	html := string(raw)

	// Prefer injecting before </body> so the snippet runs last.
	if idx := strings.LastIndex(html, "</body>"); idx != -1 {
		html = html[:idx] + liveReloadSnippet + html[idx:]
	} else {
		// Fallback: append to the end for minimal HTML files.
		html += liveReloadSnippet
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, html)
}

// handleSSE upgrades the HTTP connection to an SSE stream. The handler blocks
// until the client disconnects (r.Context() is cancelled). While connected, it
// forwards any reload signal from the sseClients registry.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE-required headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx proxy buffering.
	w.WriteHeader(http.StatusOK)

	// Register a dedicated channel for this client.
	ch := make(chan string, 1)
	s.mu.Lock()
	s.sseClients[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.sseClients, ch)
		s.mu.Unlock()
	}()

	// Send an initial "connected" comment to confirm the stream is open.
	fmt.Fprint(w, ": snapdev connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected.
			return

		case msg := <-ch:
			// Write an SSE message frame and flush immediately.
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
