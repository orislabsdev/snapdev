// Package logger provides a lightweight, colored, structured logger
// used consistently across all snapdev subsystems.
package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// ANSI escape codes used for terminal colorization.
// These are automatically stripped on non-TTY outputs in the future; for now
// they are always emitted, which is appropriate for interactive terminal use.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Level represents the severity of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelSuccess
	LevelWarn
	LevelError
)

// Logger is a simple structured logger. It writes formatted, colored lines to
// an io.Writer (typically os.Stdout / os.Stderr). It is safe to use from
// multiple goroutines because each Write call is a single fmt.Fprintln.
type Logger struct {
	out      io.Writer
	errOut   io.Writer
	minLevel Level
}

// New returns a Logger that writes info/debug/success/warn to stdout and
// errors to stderr, with LevelInfo as the minimum visible level.
func New() *Logger {
	return &Logger{
		out:      os.Stdout,
		errOut:   os.Stderr,
		minLevel: LevelInfo,
	}
}

// WithLevel returns a copy of the Logger with the minimum level adjusted.
// Use LevelDebug to see verbose output.
func (l *Logger) WithLevel(level Level) *Logger {
	return &Logger{out: l.out, errOut: l.errOut, minLevel: level}
}

// timestamp returns the current wall-clock time formatted for log lines.
func timestamp() string {
	return time.Now().Format("15:04:05")
}

// line assembles a single log line:
//
//	HH:MM:SS [LABEL] message
func line(labelColor, label, msg string) string {
	return fmt.Sprintf("%s%s%s %s%s[%s]%s %s",
		colorGray, timestamp(), colorReset,
		colorBold, labelColor, label, colorReset,
		msg,
	)
}

// Debug emits a debug-level message. Only visible when minLevel <= LevelDebug.
func (l *Logger) Debug(format string, args ...any) {
	if l.minLevel > LevelDebug {
		return
	}
	fmt.Fprintln(l.out, line(colorGray, "DBG ", fmt.Sprintf(format, args...)))
}

// Info emits a general informational message.
func (l *Logger) Info(format string, args ...any) {
	if l.minLevel > LevelInfo {
		return
	}
	fmt.Fprintln(l.out, line(colorBlue, "INFO", fmt.Sprintf(format, args...)))
}

// Success emits a green success message (e.g. after a successful build).
func (l *Logger) Success(format string, args ...any) {
	if l.minLevel > LevelSuccess {
		return
	}
	fmt.Fprintln(l.out, line(colorGreen, " OK ", fmt.Sprintf(format, args...)))
}

// Warn emits a yellow warning message for non-fatal issues.
func (l *Logger) Warn(format string, args ...any) {
	if l.minLevel > LevelWarn {
		return
	}
	fmt.Fprintln(l.out, line(colorYellow, "WARN", fmt.Sprintf(format, args...)))
}

// Error emits a red error message to stderr.
func (l *Logger) Error(format string, args ...any) {
	fmt.Fprintln(l.errOut, line(colorRed, "ERR ", fmt.Sprintf(format, args...)))
}

// Build emits a cyan build-phase message (e.g. "running npm run build").
func (l *Logger) Build(format string, args ...any) {
	if l.minLevel > LevelInfo {
		return
	}
	fmt.Fprintln(l.out, line(colorCyan, "BLD ", fmt.Sprintf(format, args...)))
}

// Banner prints the snapdev startup banner.
func (l *Logger) Banner(version string) {
	fmt.Fprintln(l.out, colorCyan+colorBold+`
  ___ _ __   __ _ _ __   __| | _____   __
 / __| '_ \ / _' | '_ \ / _' |/ _ \ \ / /
 \__ \ | | | (_| | |_) | (_| |  __/\ V /
 |___/_| |_|\__,_| .__/ \__,_|\___| \_/
                 |_|`+colorReset)
	fmt.Fprintf(l.out, "  %sLightweight build-watch-serve for React/Vite%s  v%s\n\n",
		colorGray, colorReset, version)
}
