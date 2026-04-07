// Package cmd defines the snapdev CLI commands using Cobra.
//
// The root command is the primary entry-point: it loads config, validates it,
// runs an initial build, starts the HTTP server, and enters the file-watch loop.
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/orislabsdev/snapdev/builder"
	"github.com/orislabsdev/snapdev/config"
	"github.com/orislabsdev/snapdev/logger"
	"github.com/orislabsdev/snapdev/server"
	"github.com/orislabsdev/snapdev/watcher"
	"github.com/spf13/cobra"
)

// Version is injected at build time via -ldflags:
//
//	go build -ldflags "-X github.com/orislabsdev/snapdev/cmd.Version=1.2.3"
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var (
	// CLI flag values — bound in init(), read in runRoot.
	flagConfig      string
	flagWatch       string
	flagOutput      string
	flagBuild       string
	flagPort        int
	flagHost        string
	flagProxy       string
	flagNoReload    bool
	flagVerbose     bool
	flagInitialOnly bool
)

// rootCmd is the top-level Cobra command for `snapdev`.
var rootCmd = &cobra.Command{
	Use:   "snapdev",
	Short: "Lightweight build-watch-serve for React/Vite projects",
	Long: `snapdev watches your source files, compiles your project on every
detected change, and serves the resulting static output via a minimal
HTTP server — with optional browser live reload.

It is designed as a low-memory alternative to running the Vite development
server. Instead of keeping the entire module graph in RAM, snapdev delegates
compilation to your existing build command (npm run build, vite build, etc.)
and only keeps a tiny HTTP process alive.

Quick start:

  snapdev                                # uses snapdev.json or defaults
  snapdev -w src -o dist -b "npm run build" -p 3000
  snapdev --config my-snapdev.json

See https://github.com/orislabsdev/snapdev for full documentation.`,
	SilenceUsage: true,
	RunE:         runRoot,
}

// Execute is called by main.go and runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags are available to all sub-commands (root + version).
	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "c", "snapdev.json",
		"Path to snapdev JSON config file")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false,
		"Enable debug-level log output")

	// Flags specific to the root (start) command.
	rootCmd.Flags().StringVarP(&flagWatch, "watch", "w", "",
		"Directory to watch for changes (overrides config)")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "",
		"Directory containing compiled static files (overrides config)")
	rootCmd.Flags().StringVarP(&flagBuild, "build", "b", "",
		"Build command to run on changes (overrides config)")
	rootCmd.Flags().IntVarP(&flagPort, "port", "p", 0,
		"Port for the static file server (overrides config)")
	rootCmd.Flags().StringVar(&flagHost, "host", "",
		"Host/address to bind the server to (overrides config)")
	rootCmd.Flags().StringVarP(&flagProxy, "proxy", "P", "",
		"Target URL for the reverse proxy (overrides config)")
	rootCmd.Flags().BoolVar(&flagNoReload, "no-live-reload", false,
		"Disable automatic browser live reload after builds")
	rootCmd.Flags().BoolVar(&flagInitialOnly, "build-only", false,
		"Run the initial build then exit (no server or watcher)")

	// Register sub-commands.
	rootCmd.AddCommand(versionCmd)
}

// runRoot is the main execution function for the root command.
func runRoot(cmd *cobra.Command, args []string) error {
	log := logger.New()
	if flagVerbose {
		log = log.WithLevel(logger.LevelDebug)
	}

	log.Banner(Version)

	// ── 1. Load and merge configuration ─────────────────────────────────────

	cfg, err := config.LoadFromFile(flagConfig)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// CLI flags override file config.
	if flagWatch != "" {
		cfg.WatchDir = flagWatch
	}
	if flagOutput != "" {
		cfg.OutputDir = flagOutput
	}
	if flagBuild != "" {
		cfg.BuildCommand = flagBuild
	}
	if flagPort != 0 {
		cfg.Port = flagPort
	}
	if flagHost != "" {
		cfg.Host = flagHost
	}
	if flagProxy != "" {
		cfg.ReverseProxy = flagProxy
	}
	if flagNoReload {
		cfg.LiveReload = false
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	log.Info("watch=%s  build=%q  output=%s  port=%d  live-reload=%v  proxy=%q",
		cfg.WatchDir, cfg.BuildCommand, cfg.OutputDir, cfg.Port, cfg.LiveReload, cfg.ReverseProxy)

	// ── 2. Initial build ─────────────────────────────────────────────────────

	b := builder.New(cfg, log)

	log.Build("Running initial build…")
	result := b.Run()
	if result.Success {
		log.Success("Initial build completed in %s", result.Duration.Round(time.Millisecond))
	} else {
		// A failed initial build is not fatal — the developer may fix it
		// without restarting snapdev.
		log.Warn("Initial build failed — fix the errors above and save to retry")
	}

	// --build-only exits after the first build, useful in CI pipelines.
	if flagInitialOnly {
		if !result.Success {
			return fmt.Errorf("build failed")
		}
		return nil
	}

	// ── 3. Start HTTP server ─────────────────────────────────────────────────

	srv := server.New(cfg, log)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	// ── 4. Start file watcher ─────────────────────────────────────────────────

	w, err := watcher.New(cfg, log)
	if err != nil {
		return fmt.Errorf("watcher initialisation failed: %w", err)
	}
	if err := w.Start(); err != nil {
		return fmt.Errorf("watcher failed to start: %w", err)
	}

	// ── 5. Event loop ─────────────────────────────────────────────────────────

	go func() {
		for event := range w.Events {
			log.Info("Change detected: %s [%s]", event.Path, event.Operation)

			res := b.Run()
			if res.Success {
				log.Success("Build finished in %s", res.Duration.Round(time.Millisecond))
				if cfg.LiveReload {
					srv.NotifyReload()
				}
			}
			// On failure, the builder already logged the error.
		}
	}()

	// ── 6. Graceful shutdown ──────────────────────────────────────────────────

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("Received %s — shutting down…", sig)
	w.Stop()
	srv.Shutdown()
	log.Info("Goodbye.")
	return nil
}
