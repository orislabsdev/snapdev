# Release Notes - v0.1.0

We are excited to announce the first official release of **snapdev**! 🚀

`snapdev` is a lightweight, high-performance development tool designed as a memory-efficient alternative to full-featured dev servers like Vite. It focuses on the core workflow: **watch, build, and serve.**

## Key Features

- **Low Memory Footprint**: Uses ~10MB of RAM, compared to 200MB+ for traditional dev servers.
- **Tool-Agnostic**: Works with any build command (`npm run build`, `vite build`, `hugo`, etc.).
- **Live Reload**: Automatically reloads your browser after every successful build via SSE.
- **SPA Ready**: Built-in support for single-page applications with transparent `index.html` fallback.
- **Zero Config**: Sensible defaults allow you to start with just `snapdev`, or customize via `snapdev.json`.

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

This initial release establishes a solid foundation for lightweight frontend development. We plan to add:
- More granular watch/ignore filters.
- Support for custom SSE events.
- Enhanced telemetry for build performance.

Thank you for trying out `snapdev`!

---
[MIT License](LICENSE) | [GitHub Repository](https://github.com/orislabsdev/snapdev)
