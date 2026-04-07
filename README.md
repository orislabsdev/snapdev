# snapdev

> **Build-watch-serve** — a lightweight dev tool for React/Vite that watches your source files, compiles on every change, and serves the static output with live reload.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/orislabsdev/snapdev/go.yml?branch=master&label=CI&logo=github)](https://github.com/orislabsdev/snapdev/actions/workflows/go.yml)

---

## Why snapdev?

The standard `vite` dev server keeps your **entire module graph in memory** — great for instant HMR, but expensive on constrained machines (CI runners, older laptops, Docker containers with memory limits). For many workflows — static sites, component libraries, embedded targets — you only need:

1. **Watch** source files for changes.
2. **Build** the project with your normal production build command.
3. **Serve** the resulting static files with automatic browser reload.

`snapdev` does exactly this in a single ~10 MB binary with near-zero idle memory.

| Feature | Vite dev server | snapdev |
|---|---|---|
| Hot Module Replacement | ✅ Full HMR | ❌ Full-page reload |
| Idle memory usage | ~200–500 MB | ~5–15 MB |
| Incremental builds | ✅ In-memory | ✅ Via your build tool |
| SPA fallback routing | ✅ | ✅ |
| Live browser reload | ✅ | ✅ (SSE) |
| Works with any bundler | ❌ Vite-specific | ✅ Any command |
| Single binary / no Node | ❌ | ✅ |

---

## Installation

### Pre-built binary (recommended)

Download the latest release for your platform from the [Releases page](https://github.com/orislabsdev/snapdev/releases).

```bash
# macOS / Linux — place in your PATH
curl -sSL https://github.com/orislabsdev/snapdev/releases/latest/download/snapdev-$(uname -s | tr A-Z a-z)-amd64.tar.gz | tar -xz
sudo mv snapdev /usr/local/bin/
```

### Build from source

```bash
git clone https://github.com/orislabsdev/snapdev.git
cd snapdev
make build          # outputs ./bin/snapdev
make install        # installs to $GOPATH/bin
```

Requires **Go 1.21+**.

### go install

```bash
go install github.com/orislabsdev/snapdev@latest
```

---

## Quick start

```bash
# In your React/Vite project directory:
snapdev

# Or with explicit flags:
snapdev --watch src --output dist --build "npm run build" --port 3000
```

Open **http://localhost:3000**. Save a file in `src/` — the project rebuilds and the browser reloads automatically.

---

## Configuration

`snapdev` looks for a `snapdev.json` file in the current directory. All fields are optional and fall back to sensible defaults.

```json
{
  "watchDir":     "src",
  "buildCommand": "npm run build",
  "outputDir":    "dist",
  "port":         3000,
  "host":         "localhost",
  "debounceMs":   300,
  "liveReload":   true,
  "ignore":       ["node_modules", ".git", "dist", ".snapdev"],
  "extensions":   [".tsx", ".ts", ".jsx", ".js", ".css", ".html", ".json", ".svg"]
}
```

### Configuration reference

| Field | Type | Default | Description |
|---|---|---|---|
| `watchDir` | string | `"src"` | Root directory to watch for changes |
| `buildCommand` | string | `"npm run build"` | Shell command executed on each change |
| `outputDir` | string | `"dist"` | Directory of compiled assets to serve |
| `port` | integer | `3000` | HTTP port |
| `host` | string | `"localhost"` | Bind address (`"0.0.0.0"` to expose externally) |
| `debounceMs` | integer | `300` | Milliseconds of quiet time before triggering a build |
| `liveReload` | boolean | `true` | Inject SSE snippet and reload browser after builds |
| `ignore` | string[] | `[…]` | Path substrings to exclude from watching |
| `extensions` | string[] | `[…]` | File extensions that trigger a rebuild |

### CLI flags

Flags always override `snapdev.json`:

```
  -c, --config string         Path to config file (default "snapdev.json")
  -w, --watch string          Directory to watch
  -o, --output string         Directory to serve
  -b, --build string          Build command
  -p, --port int              HTTP port
      --host string           Bind address
      --no-live-reload        Disable live reload
      --build-only            Run one build then exit (useful in CI)
  -v, --verbose               Debug-level logging
  -h, --help                  Show help
```

---

## Usage examples

### Vite project (default config)

```bash
# snapdev.json is optional — all of these are the defaults:
snapdev
```

### Custom build command

```bash
snapdev --build "pnpm run build:prod"
```

### Multiple projects (different ports)

```bash
snapdev --config apps/admin/snapdev.json &
snapdev --config apps/dashboard/snapdev.json --port 3001 &
```

### Docker

```dockerfile
FROM golang:1.21 AS snapdev-builder
RUN go install github.com/orislabsdev/snapdev@latest

FROM node:20-slim
COPY --from=snapdev-builder /go/bin/snapdev /usr/local/bin/snapdev
WORKDIR /app
COPY . .
RUN npm ci
EXPOSE 3000
CMD ["snapdev", "--host", "0.0.0.0"]
```

### CI — build check only

```bash
snapdev --build-only     # exits 0 on success, 1 on failure
```

---

## How it works

```
┌──────────────┐   file change    ┌─────────────────┐   success   ┌─────────────────┐
│  fsnotify    │ ──(debounced)──► │  build command  │ ──────────► │   SSE broadcast │
│  recursive   │                  │  (your bundler) │             │   → browser     │
│  watcher     │                  └─────────────────┘             │     reload      │
└──────────────┘                                                   └─────────────────┘
                                                                           │
                                                    ┌──────────────────────┘
                                                    │
                                             ┌──────▼──────┐
                                             │  net/http   │
                                             │  SPA server │
                                             │  :3000      │
                                             └─────────────┘
```

1. **Watcher** — `fsnotify` monitors the watch directory recursively. Events are debounced so that multiple rapid saves produce a single build trigger.
2. **Builder** — runs your build command in a subprocess (`sh -c` on Unix, `cmd /C` on Windows). Stdout/stderr are captured and surfaced in the snapdev log on failure.
3. **Server** — `net/http` serves files from the output directory. Unknown paths fall back to `index.html` for SPA client-side routing.
4. **Live reload** — when enabled, a tiny `<script>` is injected into HTML responses that opens an SSE connection to `/__snapdev_sse`. After each successful build, snapdev sends a `reload` event and every connected tab refreshes.

---

## Supported bundlers

`snapdev` delegates building to whatever command you configure, so it works with any bundler or static site generator:

| Tool | `buildCommand` example |
|---|---|
| Vite | `"npm run build"` or `"vite build"` |
| Create React App | `"react-scripts build"` |
| Parcel | `"parcel build src/index.html"` |
| Webpack | `"webpack --mode production"` |
| Next.js (export) | `"next build && next export"` |
| Astro | `"astro build"` |
| SvelteKit | `"vite build"` |
| Hugo | `"hugo"` |

---

## Development

```bash
git clone https://github.com/orislabsdev/snapdev.git
cd snapdev

make deps          # download Go modules
make build         # build ./bin/snapdev
make test          # run the test suite
make lint          # run golangci-lint
make build-all     # cross-compile for all platforms
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full contribution guidelines.

---

## Project structure

```
snapdev/
├── cmd/
│   ├── root.go          # Root Cobra command — wires all subsystems
│   └── version.go       # `snapdev version` sub-command
├── builder/
│   └── builder.go       # Subprocess build executor
├── config/
│   └── config.go        # Config loading, defaults, validation
├── logger/
│   └── logger.go        # Coloured, levelled logger
├── server/
│   └── server.go        # Static file server + SSE live reload
├── watcher/
│   └── watcher.go       # fsnotify-backed file watcher with debounce
├── main.go
├── go.mod
├── Makefile
├── snapdev.json         # Example configuration
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
└── LICENSE
```

---

## License

[MIT](LICENSE) &copy; 2026 Oris Labs. Built by engineers, for engineers.
