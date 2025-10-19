# Goobergine Application Generator

Goobergine is a Go-based application generator designed to scaffold and manage self-contained Go web applications with clean interfaces, predictable structure, and batteries-included defaults.

## Overview

The goal of **Goobergine v0.1-alpha** is to produce a fully functional backend skeleton with:

- A Go web server using `modernc.org/sqlite` as the datastore.
- An HTMX + AlpineJS + missing.style frontend.
- A local-only admin API for safe management.
- Graceful shutdowns, versioned schema management, and clear configuration rules.

Future versions (starting with **v0.2**) will add frontend UX elements such as the login page, dashboard templates, and user session management UI.

## Quickstart

```bash
# Create and initialize the store
app db create

# Start the server (defaults to ports 8080 and 8383)
app serve --port 8080 --admin-port 8383

# Check server status (via local admin API)
app server status

# Perform a graceful restart
app server restart

# Shut down the server
app server shutdown
```

## Directory Layout

```
cmd/{{.BinaryName}}/main.go
internal/server/
internal/admin/
internal/session/
internal/store/
templates/
README.md
SECURITY_CONSIDERATIONS.md
```

## Design Philosophy

- **Tight interfaces, replaceable implementations.**  
  Each subsystem defines a Goob contract — small, stable, testable interfaces with reference implementations.

- **Single binary simplicity.**  
  Both the web server and admin CLI live in one executable for minimal deployment friction.

- **Portable defaults.**  
  Works on Windows, macOS, and Linux without requiring CGO.

- **Predictable behavior.**  
  No magic, no globals. Startup and shutdown are explicit and logged.

## Configuration

Configuration follows this precedence order:

1. Built-in defaults  
2. Values in the `app_config` table  
3. Command-line flags or environment variables  
4. One-time overrides during `app db create`

Example:

```bash
app db create --session-idle=30m --session-abs=24h
app db update sessions session-idle 45m
```

All admin operations use the JSON-only API on the loopback interface.

## Frontend (v0.1)

The frontend is intentionally minimal:

- HTML fragments rendered by the Go server.
- Enhanced by HTMX and AlpineJS.
- Styled with missing.style.
- Installation page serves when the store is missing or under maintenance.

## Security Highlights

- Admin API loopback-only (`127.0.0.1`, `::1`).
- JSON-only admin routes.
- Session cookies with secure attributes.
- CSRF protection via maintained middleware.
- No remote administration in v0.1.

## Roadmap

- **v0.2:** Frontend UX (login page, dashboard, form actions)
- **v0.3:** Configuration editor via admin UI
- **v0.4:** Role-based access controls (RBAC)
- **v1.0:** Hardened security, remote admin, full testing suite

## Acknowledgment

The initial design, structure, and TODO documentation for Goobergine were collaboratively created using **ChatGPT (OpenAI)** to accelerate architecture drafting and consistency across components.

## License

MIT — see `LICENSE` file when available.

---
From tabula rasa to orbis terrarum in minutes.
