// Package config defines and loads snapdev configuration from a JSON file
// or sensible defaults, with support for CLI flag overrides.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds all runtime configuration for snapdev.
// Values can be set via snapdev.json or overridden through CLI flags.
type Config struct {
	// WatchDir is the root directory to monitor for file changes (e.g. "src").
	WatchDir string `json:"watchDir"`

	// BuildCommand is the shell command executed on each detected change (e.g. "npm run build").
	BuildCommand string `json:"buildCommand"`

	// OutputDir is the directory that receives the compiled static assets (e.g. "dist").
	OutputDir string `json:"outputDir"`

	// Port is the TCP port on which the static file server listens.
	Port int `json:"port"`

	// Debounce is the quiet period after the last file event before a build is triggered.
	// This prevents multiple rapid saves from launching redundant builds.
	// Stored as milliseconds in JSON (e.g. 300).
	Debounce time.Duration `json:"-"`

	// DebounceMs is the JSON-facing field for Debounce (in milliseconds).
	DebounceMs int `json:"debounceMs"`

	// Ignore is a list of path substrings to exclude from watching.
	// Common entries: "node_modules", ".git", "dist".
	Ignore []string `json:"ignore"`

	// LiveReload controls whether a browser reload is triggered after each successful build.
	LiveReload bool `json:"liveReload"`

	// Extensions is the list of file extensions that will trigger a rebuild when changed.
	Extensions []string `json:"extensions"`

	// Host is the address the server binds to. Defaults to "localhost".
	// Set to "0.0.0.0" to expose on all interfaces (e.g. within Docker).
	Host string `json:"host"`

	// ReverseProxy is the optional target URL for the reverse proxy.
	// If set, requests that don't match a local file will be forwarded here.
	ReverseProxy string `json:"reverseProxy"`
}

// DefaultConfig returns a Config populated with sensible defaults for a
// typical Vite/React project. The caller may override individual fields.
func DefaultConfig() *Config {
	return &Config{
		WatchDir:     "src",
		BuildCommand: "npm run build",
		OutputDir:    "dist",
		Port:         3000,
		Host:         "localhost",
		DebounceMs:   300,
		Debounce:     300 * time.Millisecond,
		Ignore:       []string{"node_modules", ".git", "dist", ".snapdev"},
		LiveReload:   true,
		Extensions: []string{
			".tsx", ".ts", ".jsx", ".js",
			".css", ".scss", ".sass", ".less",
			".html", ".json", ".svg", ".env",
		},
	}
}

// LoadFromFile reads a JSON configuration file at path and merges its values
// on top of DefaultConfig. If the file does not exist, the defaults are
// returned without error — this allows snapdev to run with zero config.
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file is perfectly fine; use defaults.
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	// Convert the JSON-friendly millisecond integer into a time.Duration.
	if cfg.DebounceMs > 0 {
		cfg.Debounce = time.Duration(cfg.DebounceMs) * time.Millisecond
	}

	return cfg, nil
}

// Validate checks that the required Config fields are non-empty and returns
// an error describing the first problem found.
func (c *Config) Validate() error {
	if c.WatchDir == "" {
		return fmt.Errorf("watchDir must not be empty")
	}
	if c.BuildCommand == "" {
		return fmt.Errorf("buildCommand must not be empty")
	}
	if c.OutputDir == "" {
		return fmt.Errorf("outputDir must not be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port %d is out of range (1–65535)", c.Port)
	}
	return nil
}
