// Package builder executes the configured build command (e.g. "npm run build")
// in a subprocess, captures its output, and returns a structured result.
package builder

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/orislabsdev/snapdev/config"
	"github.com/orislabsdev/snapdev/logger"
)

// Result contains the outcome of a single build run.
type Result struct {
	// Success is true when the build command exited with code 0.
	Success bool
	// Duration is the wall-clock time taken by the build process.
	Duration time.Duration
	// Stdout captures all bytes written to the subprocess's standard output.
	Stdout string
	// Stderr captures all bytes written to the subprocess's standard error.
	// On failure this typically contains the compiler or bundler error message.
	Stderr string
}

// Builder executes build commands and reports results.
// It is intentionally stateless so that concurrent calls are safe, though in
// practice snapdev serialises builds through its event loop.
type Builder struct {
	cfg *config.Config
	log *logger.Logger
}

// New returns a Builder configured with the given cfg and log.
func New(cfg *config.Config, log *logger.Logger) *Builder {
	return &Builder{cfg: cfg, log: log}
}

// Run executes cfg.BuildCommand as a subprocess, waits for it to exit, and
// returns a Result. The build's stdout/stderr are captured — not streamed to
// the terminal — so that snapdev's own log output stays readable. On failure
// the error output is printed via the logger.
func (b *Builder) Run() Result {
	start := time.Now()
	b.log.Build("Running: %s", b.cfg.BuildCommand)

	cmd, err := parseCommand(b.cfg.BuildCommand)
	if err != nil {
		return Result{
			Success:  false,
			Duration: time.Since(start),
			Stderr:   err.Error(),
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	elapsed := time.Since(start)

	result := Result{
		Success:  runErr == nil,
		Duration: elapsed,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
	}

	if !result.Success {
		// Print the build error so the developer can act on it without leaving
		// their terminal to look at a separate build window.
		errMsg := result.Stderr
		if errMsg == "" {
			errMsg = fmt.Sprintf("process exited with: %v", runErr)
		}
		b.log.Error("Build failed after %s:\n%s", elapsed.Round(time.Millisecond), errMsg)
	}

	return result
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// parseCommand splits a shell command string into its executable and arguments,
// accounting for Windows (cmd /C ...) vs Unix (/bin/sh -c ...) differences.
//
// Using the shell as the executor instead of exec.Command(parts[0], parts[1:]...)
// allows the command string to contain shell features like environment variable
// expansion, pipes, and && chaining — exactly what developers expect.
func parseCommand(command string) (*exec.Cmd, error) {
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("build command is empty")
	}

	if runtime.GOOS == "windows" {
		// On Windows, route through cmd.exe so that npm.cmd scripts work.
		return exec.Command("cmd", "/C", command), nil
	}

	// On Unix, route through sh so that environment variables, $PATH lookup,
	// and compound commands (&&, ||, ;) all work as expected.
	return exec.Command("sh", "-c", command), nil
}