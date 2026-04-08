# Release Notes - v0.2.1

We are pleased to announce version **v0.2.1** of **snapdev**! 🚀

This release focuses on refining the **Reverse Proxy** feature introduced in v0.2.0, adding comprehensive **unit tests**, and improving the **SPA routing** logic to better handle browser navigation.

## What's New?

### 🛠️ Smart SPA Routing
We've improved how `snapdev` decides whether to serve your `index.html` or forward a request to the reverse proxy. 

Browsers performing full-page navigations (which include `Accept: text/html` in their headers) now correctly receive the `index.html` fallback for unknown routes. This ensures that client-side routing (like React Router) works seamlessly even when a proxy is configured for API calls.

### 🔌 Improved Proxy Compatibility
The reverse proxy now correctly propagates the `Host` header to the target backend. This is a critical fix for users proxying to servers that rely on virtual hosting or specific host-based middleware.

### 🧪 Enhanced Test Suite
We've added a comprehensive unit test suite for the `server` package, covering static file resolution, SPA fallback, and proxying logic. This ensures a stable and predictable developer experience as the project grows.

### ✨ Visual Polish
- Refined the startup banner for a cleaner look.
- Improved code formatting and alignment in tests.

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

Our roadmap continues:
- **Watch Filters**: More granular control over file inclusions and exclusions.
- **Custom Events**: Support for user-defined SSE events.
- **Plugin System**: Research into custom build hooks.

Thank you for your feedback and contributions!

---
[MIT License](LICENSE) | [GitHub Repository](https://github.com/orislabsdev/snapdev)
