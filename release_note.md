# Release Notes - v0.3.0

We are excited to announce version **v0.3.0** of **snapdev**! 🚀

This is a major feature release that introduces **CSS Hot Module Replacement (HMR)**, allowing you to see style changes instantly without losing your application state or refreshing the page.

## What's New?

### 🔥 CSS Hot Module Replacement (HMR)
Say goodbye to full-page refreshes when you're just tweaking margins or colors! 

`snapdev` now distinguishes between style changes and logic changes. When you modify a **CSS, SCSS, Sass, or Less** file, `snapdev` sends a targeted update signal to the browser. The injected live-reload snippet then hot-swaps your stylesheets in real-time.

*   **Fast**: No waiting for the page to reload and re-render.
*   **State-preserving**: Keep your form inputs, scroll position, and UI state intact while styling.
*   **Agnostic**: Works with any bundler (Vite, Webpack, Parcel, etc.) as long as it outputs standard CSS.

### 🏗️ Targeted SSE Notifications
We've refactored our internal notification system to support more than just "reload". This paves the way for future HMR features for images, assets, and even JavaScript modules.

### ✨ Other Improvements
- Bumped version to **v0.3.0**.
- Cleaned up the injected live-reload script for better reliability and performance.
- Improved logging for development builds.

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
- **Asset HMR**: Hot-swapping images and other assets.
- **Plugin System**: Research into custom build hooks.

Thank you for your feedback and contributions!

---
[MIT License](LICENSE) | [GitHub Repository](https://github.com/orislabsdev/snapdev)
