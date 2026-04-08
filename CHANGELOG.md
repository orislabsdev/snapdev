# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-04-08

### Added
- Reverse Proxy support: automatically forward requests to a backend server when local files are not found (e.g., API requests).
- New CLI flag `--proxy` (or `-P`) and `reverseProxy` configuration field in `snapdev.json`.
- Integrated GoReleaser for automated multi-platform binary distribution (Linux, macOS, Windows).
- New `release` targets in Makefile for streamlined local release testing.
- Enhanced version command with build metadata (Commit, Build Date) via GoReleaser.

### Changed
- Upgraded project to Go 1.25.1 for performance and security improvements.
- Improved Makefile compatibility for macOS (`mkdir` flag fixes).
- Simplified pre-built binary installation instructions in documentation.
- Refactored server logic to support proxy-then-SPA fallback mechanism.
- Disabled buildvcs in development builds for faster iteration.

## [0.1.0] - 2026-04-07

### Added
- Initial release of `snapdev`.
- Recursive file watcher with `fsnotify` and automatic debouncing.
- Subprocess build executor supporting any CLI command.
- Static file server with SPA routing fallback (index.html).
- Live reload via Server-Sent Events (SSE) with automatic script injection.
- JSON configuration support (`snapdev.json`) and CLI flag overrides.
- Multi-platform support (Linux, macOS, Windows).
- GitHub Actions CI/CD for cross-compilation, linting, and testing.
- Documentation: README, CONTRIBUTING, SECURITY, and code documentation.
- Added build metadata to version command.

### Changed
- Standardized logging implementation with colored output and level control.
- Refactored internal packages to root and updated Go version requirement to 1.25.1.
- Updated repository URLs and module path to `orislabsdev/snapdev`.
- Improved configuration management and validation.

### Fixed
- Various minor improvements to error handling and watcher stability.
- Makefile: removed `-v` flag from `mkdir` to ensure compatibility with macOS `mkdir`.
