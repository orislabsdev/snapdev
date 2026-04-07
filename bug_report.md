---
name: Bug report
about: Something isn't working as expected
title: "bug: "
labels: bug
assignees: ""
---

## Describe the bug

A clear and concise description of what the bug is.

## Steps to reproduce

1. Run `snapdev --version`
2. Run `snapdev -w src -b "npm run build" -o dist`
3. Save a file in `src/`
4. See error…

## Expected behaviour

What you expected to happen.

## Actual behaviour

What actually happened. Please include the full terminal output with `--verbose` enabled.

```
paste output here
```

## Environment

| Field | Value |
|---|---|
| snapdev version | `snapdev version` output |
| OS | e.g. macOS 14.3, Ubuntu 22.04, Windows 11 |
| Go version | `go version` |
| Node.js version | `node --version` |
| Bundler / build command | e.g. `npm run build` (Vite 5.0) |

## Additional context

Any other information, config files (`snapdev.json`), or screenshots.