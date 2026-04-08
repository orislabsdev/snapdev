# Release Notes - v0.2.0

We are excited to announce version **v0.2.0** of **snapdev**! 🚀

This release introduces highly requested features, including **Reverse Proxy support** and automated **multi-platform distribution** via GoReleaser.

## What's New?

### 🔄 Reverse Proxy Support
You can now redirect API requests or other non-static assets to a different backend server. If a requested file is not found in the `outputDir`, `snapdev` will forward the request to your configured proxy target.

- **CLI**: `snapdev --proxy http://localhost:8080`
- **JSON**: `"reverseProxy": "http://localhost:8080"`

### 📦 Automated Releases
We've integrated **GoReleaser** to provide verified, cross-platform binaries for Linux, macOS, and Windows. This ensures that you always have access to the latest performance improvements and security patches on any system.

### ⚙️ Performance & Stability
- **Go 1.25.1**: Upgraded the project to the latest Go version for improved runtime performance.
- **Optimised Builds**: Disabled unnecessary VCS metadata collection during development builds to speed up the watch-rebuild loop.
- **macOS Compatibility**: Fixed Makefile issues for better developer experience on Apple Silicon and Intel Macs.

## Installation

### Pre-built binary

```bash
curl -sSL https://github.com/orislabsdev/snapdev/releases/latest/download/snapdev-$(uname -s | tr A-Z a-z)-amd64.tar.gz | tar -xz
sudo mv snapdev /usr/local/bin/
```

### Go install

```bash
go install github.com/orislabsdev/snapdev@latest
```

## What's Next?

Our roadmap for the upcoming releases includes:
- **Watch Filters**: More granular control over file inclusions and exclusions.
- **Custom Events**: Support for user-defined SSE events for complex frontend interactions.
- **Plugin System**: Initial research into a lightweight plugin architecture for custom build steps.

Thank you for being part of the `snapdev` journey!

---
[MIT License](LICENSE) | [GitHub Repository](https://github.com/orislabsdev/snapdev)
