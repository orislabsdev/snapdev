# Contributing to snapdev

Thank you for your interest in contributing! This document covers everything you need to get started.

## Table of contents

- [Code of conduct](#code-of-conduct)
- [Ways to contribute](#ways-to-contribute)
- [Development setup](#development-setup)
- [Project conventions](#project-conventions)
- [Submitting a pull request](#submitting-a-pull-request)
- [Reporting issues](#reporting-issues)

---

## Code of conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating you agree to uphold these standards. Violations may be reported to the maintainers via the contact information in [SECURITY.md](SECURITY.md).

---

## Ways to contribute

You do not need to write code to contribute. We welcome:

- **Bug reports** — detailed descriptions of unexpected behaviour with reproduction steps.
- **Feature requests** — well-scoped ideas with a clear use case.
- **Documentation improvements** — typo fixes, clearer wording, additional examples.
- **Code contributions** — bug fixes, new features, tests, or refactoring.

---

## Development setup

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.21+ | Building and testing |
| Make | any | Running development tasks |
| golangci-lint | latest | Linting (optional but recommended) |
| Node.js | 18+ | Testing with a real React/Vite project |

### Clone and build

```bash
git clone https://github.com/orislabsdev/snapdev.git
cd snapdev
make deps    # go mod download
make build   # outputs ./bin/snapdev
```

### Running in development

```bash
# Run against a local React/Vite project:
./bin/snapdev --watch ../my-react-app/src \
              --output ../my-react-app/dist \
              --build "cd ../my-react-app && npm run build" \
              --port 3000 \
              --verbose
```

### Available Make targets

```
make build        Build binary to ./bin/snapdev
make install      Install to $GOPATH/bin
make test         Run the full test suite
make test-race    Run tests with -race detector
make lint         Run golangci-lint
make vet          Run go vet
make deps         Download Go module dependencies
make tidy         go mod tidy
make build-all    Cross-compile for linux/darwin/windows (amd64 + arm64)
make clean        Remove build artefacts
make release      Tag and push a release (maintainers only)
```

---

## Project conventions

### Code style

- Follow standard Go conventions (`gofmt`, `goimports`).
- Every exported symbol (type, function, constant) must have a doc comment.
- Keep functions short and focused; prefer small, composable units.
- Avoid external dependencies unless they provide significant value — the goal is a small, auditable binary.

### Packages

| Package | Responsibility |
|---|---|
| `cmd` | CLI definition (Cobra commands, flag parsing, wiring) |
| `config` | Config struct, defaults, loading, validation |
| `logger` | Coloured terminal output |
| `watcher` | fsnotify wrapper with debounce and filtering |
| `builder` | Subprocess build execution |
| `server` | HTTP server, SPA fallback, SSE live reload |

Do not introduce cross-package imports that form cycles. All packages are independent; only `cmd` wires them together.

### Tests

- Unit tests live alongside their source files (`builder_test.go`, etc.).
- Integration tests that shell out to real processes live in `internal/**/*_integration_test.go` and are gated by `//go:build integration`.
- Test coverage should be added for any new public function.
- Use `t.TempDir()` for temporary directories; never hardcode `/tmp`.

### Commit messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <short summary>

[optional body]
[optional footer(s)]
```

Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`.

Examples:

```
feat(server): add SPA fallback routing for React Router
fix(watcher): prevent duplicate events on rapid saves
docs: add Docker usage example to README
test(builder): cover Windows cmd.exe command parsing
```

---

## Submitting a pull request

1. **Fork** the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Write tests** for any new behaviour. Run the full suite:
   ```bash
   make test
   make lint
   ```

3. **Update documentation** if your change affects public behaviour, config options, or CLI flags.

4. **Open a PR** against `main`. Fill in the pull request template completely, including:
   - What the change does and why.
   - How to test it manually.
   - Any screenshots or terminal output that illustrates the change.

5. A maintainer will review within a few business days. Please be patient and responsive to feedback.

### PR checklist

- [ ] Tests added / updated
- [ ] `make test` passes locally
- [ ] `make lint` passes (or issues are discussed in the PR)
- [ ] Documentation updated (README, config reference, etc.)
- [ ] Commit messages follow Conventional Commits

---

## Reporting issues

Use the GitHub [issue tracker](https://github.com/orislabsdev/snapdev/issues).

- Search for existing issues before opening a new one.
- Use the **Bug Report** template for bugs and the **Feature Request** template for ideas.
- Include your OS, Go version, Node.js version, and the `snapdev version` output.
- Attach the full terminal output with `--verbose` enabled if relevant.

For security vulnerabilities, please **do not** open a public issue — see [SECURITY.md](SECURITY.md) instead.