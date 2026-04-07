# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| latest (`main`) | ✅ Active security support |
| older releases | ❌ Please upgrade |

We recommend always using the latest release.

---

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues, pull requests, or discussions.**

If you believe you have found a security vulnerability in snapdev, please report it responsibly using one of the following methods:

### Option A — GitHub private vulnerability reporting (preferred)

1. Go to the repository's **Security** tab.
2. Click **"Report a vulnerability"**.
3. Fill in the form with as much detail as possible.

GitHub will notify the maintainers privately and we will coordinate disclosure.

### Option B — Email

Send a description of the issue to **security@snapdev.dev** (replace with real address before publishing).

Please include:

- A concise description of the vulnerability.
- Steps to reproduce (proof-of-concept if available).
- The potential impact and any mitigations you are aware of.
- Your name / handle if you would like credit.

You will receive an acknowledgement within **48 hours** and a resolution plan within **7 business days**.

---

## Responsible disclosure

We ask that you:

- Give us reasonable time to investigate and patch before public disclosure.
- Avoid accessing or modifying data that does not belong to you.
- Do not disrupt production systems.

In return, we will:

- Acknowledge your report promptly.
- Keep you informed of progress.
- Credit you in the release notes (unless you prefer anonymity).
- Not pursue legal action against good-faith security researchers.

---

## Security considerations

### What snapdev does

`snapdev` is a development tool. It:

- Runs an arbitrary shell command supplied by the user (`buildCommand`).
- Serves files from a local directory over HTTP.
- Injects a small `<script>` tag into HTML responses for live reload.

### Threat model

| Concern | Status |
|---|---|
| **Remote code execution via `buildCommand`** | By design — only the local developer controls this value. Never expose snapdev to untrusted networks. |
| **Directory traversal in file server** | Mitigated — paths are cleaned with `filepath.Clean` and confined to `outputDir`. |
| **SSE endpoint abuse** | Low risk — the SSE endpoint only emits `"reload"` events and reads nothing from clients. |
| **Binding to `0.0.0.0`** | The default bind address is `localhost`. Binding to `0.0.0.0` exposes the server on all interfaces — do this only in trusted environments (e.g. Docker internal network). |

### Recommendations

- **Never run snapdev in production.** It is a development-only tool.
- **Never expose snapdev's port to the public internet.** Use a firewall or VPN.
- **Use `localhost` binding** (the default) unless you have a specific need to expose the server to other machines.
- **Keep snapdev updated** to benefit from dependency security patches.