# AGENT.md — Goobergine Application Generator Agent

## Overview

This document defines the Goobergine Agent’s role in maintaining, building, and testing the **Goobergine Application Generator**.  
It describes the toolchain, project layout, conventions, and philosophy — emphasizing **Go standard library first** development.

---

## 🧰 Tech Stack Summary

### Core Language and Runtime
- **Language:** Go ≥ 1.23
- **Build Tools:** Go toolchain (`go build`, `go test`, `go tool template`)
- **No Node, Bun, or external build system required.**
- **Frontend stack:** HTMX, AlpineJS, and missing.style (no build step).

### Standard Library Usage
| Area | Standard Package | Purpose |
|------|------------------|----------|
| HTTP server | `net/http` | Core request handling |
| Templates | `html/template` | Rendering HTML fragments |
| Database | `database/sql` | Generic interface to SQLite backend |
| Logging | `log` / `log/slog` | Default logger via Goob logging contract |
| Embedding | `embed` | Include static assets & templates |
| Config & Flags | `flag`, `os`, `time`, `context` | CLI options and runtime configuration |

### Minimal Dependencies
| Purpose | Package | Reason |
|----------|----------|--------|
| Datastore | `modernc.org/sqlite` | Pure-Go SQLite driver (portable, CGO-free) |
| CLI | `spf13/cobra` | Command structure for `app` commands |
| CSRF | Maintained middleware | Security for HTML and JSON routes |
| Frontend | `missing.style`, `HTMX`, `AlpineJS` | Lightweight UI; no Node build chain |

---

## 🏗️ Build and Run Commands

### Build
```bash
make build
```
Compiles the single binary into `dist/app`.

### Run
```bash
make run PORT=8080 ADMIN_PORT=8383
```
Runs the web app (HTML) and admin API (JSON) on separate ports.

### Clean and Tidy
```bash
make clean
make tidy
```

### Binary Defaults
| Flag | Default | Description |
|------|----------|-------------|
| `--port` | `8080` | Public HTML listener |
| `--admin-port` | `8383` | JSON-only admin listener |
| `--shutdown-timeout` | `15s` | Graceful shutdown timeout |

### Commands
```bash
app db create
app db upgrade
app db verify
app serve --port 8080 --admin-port 8383
app server restart
app server shutdown
```

---

## 🧪 Testing Helpers

### Run all tests
```bash
make test
```

### Run race detector
```bash
go test -race ./...
```

### Session Store tests
```bash
go test ./internal/session -v
```

### Integration Tests
- Run against **both** in-memory and SQLite backends.
- Test via HTTP on ports `8080` and `8383`.
- Verify JSON-only admin responses and HTML fragments for HTMX.

---

## ⚙️ Contracts and Interfaces (Goob Philosophy)

Each major subsystem defines a **Goob contract** — a minimal, stable interface for extension.

| Subsystem | Contract | Purpose |
|------------|-----------|----------|
| Session | `SessionStore` | Manages user sessions and persistence |
| Logging | `Logger` | Abstract logging (default: Go std logger) |
| Store | `Store` | Database schema and access abstraction |
| Admin Channel | `AdminTransport` | Interface for local admin communication |

Contracts are tested with **shared test suites** to ensure interchangeability.

---

## 🧩 File and Directory Layout

```
cmd/app/main.go
internal/server/
internal/admin/
internal/session/
internal/store/
templates/
dist/
```

| Directory | Purpose |
|------------|----------|
| `cmd/app/` | CLI and server entry point |
| `internal/server/` | Public HTML routes (HTMX) |
| `internal/admin/` | Admin JSON API handlers |
| `internal/session/` | Session interfaces & implementations |
| `internal/store/` | SQLite logic & schema migrations |
| `templates/` | HTML fragments (install, login, dashboard, etc.) |

---

## 🧱 Development Principles

1. **Standard Library First**
   - Use Go’s `net/http`, `context`, `html/template`, and `encoding/json` before third-party packages.
   - Avoid frameworks or heavy abstractions unless proven necessary.

2. **Single Binary Simplicity**
   - All functionality (CLI, server, admin API) lives in one binary.

3. **Portability**
   - Works the same on Windows, macOS, and Linux.
   - SQLite backend is CGO-free and portable.

4. **Predictable Behavior**
   - Explicit startup, configuration, and graceful shutdown.
   - Log everything important; never panic without context.

5. **Explicit Configuration**
   - Config precedence: defaults → `app_config` table → flags/env → bootstrap overrides.
   - Admin updates always go through the JSON API.

---

## 🧭 Testing Philosophy

- Unit tests cover each contract and backend.
- Integration tests verify session behavior, admin endpoints, and install/maintenance modes.
- All tests must pass under `-race` and `-count=1`.
- Inject test doubles for RNG and Clock to simulate session expiry.

---

## 🔐 Security Model

- Admin listener bound to loopback only; rejects non-local binds.
- Admin API JSON-only (`Content-Type: application/json` required).
- Public app uses CSRF-protected HTML routes via maintained middleware.
- Cookies set with `Secure`, `HttpOnly`, and `SameSite=Lax` attributes.
- Graceful shutdown ensures data consistency (SQLite finalization).

---

## 🪶 Development Notes

- Configuration and sessions stored in SQLite under `app_config` and `sessions` tables.
- For testing or demos: `app db create --session-backend=memory` for ephemeral mode.
- Admin commands never modify the store directly (except `app db create`).
- Template changes are live when using `go run`; production builds use embedded assets.

---

## 🧭 Future Work

- v0.2: Frontend UX (login, session restore, dashboard)
- v0.3: Admin configuration editor
- v0.4: Role-based access control (RBAC)
- v1.0: Remote admin with mTLS or UDS

---

**Goobergine Agent Principle:**  
> “When the standard library works, use it. When it doesn’t, write the smallest contract that makes it replaceable.”
