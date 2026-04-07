// snapdev - A lightweight build-on-change dev server for React/Vite projects.
//
// Instead of running a memory-heavy dev bundler like `vite`, snapdev watches
// your source files, triggers a real build on every change, and serves the
// resulting static output via a minimal HTTP server with live reload.
//
// Usage:
//
//	snapdev [flags]
//	snapdev --watch src --output dist --build "npm run build" --port 3000
package main

import "github.com/orislabsdev/snapdev/cmd"

func main() {
	cmd.Execute()
}