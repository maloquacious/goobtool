# Goobergine Application Generator — TODO (v0.1-alpha, WIP)

## Architecture / Goals
[ ] Web front end (HTMX/HTML) with Go serving the backend.
[ ] Keep v0 on modules (start at v0.1-alpha; avoid v1 semantics).
[ ] Enforce JSON-only on admin APIs; public app returns HTML.

## Datastore
[ ] Default datastore: modernc.org/sqlite (portable).
[ ] Define storage interfaces (“Goob contracts”) so backends can be swapped later.
[ ] SQLite safe defaults: enable WAL, synchronous=NORMAL, foreign_keys=ON.

## Backend MVP (Go server)
[ ] Single binary provides both the server and admin CLI.
[ ] Command: app serve --port 8080 --admin-port 8383 (defaults: 8080 / 8383).
[ ] Primary listener serves the web application (HTML responses for HTMX).
[ ] Secondary listener serves the admin HTTP API (loopback only, JSON-only).
[ ] Graceful shutdown (allow SQLite flush/close).
[ ] --shutdown-timeout=15s default.
[ ] Timer flag to auto-shutdown for tests.
[ ] Lifecycle:
- [ ] If no store: app serve exits with guidance (“Run: app db create”).
- [ ] If store exists but wrong version/uninitialized: log & serve installation app.
- [ ] Health endpoints: /live (OK when process is up), /ready (OK only when store initialized and not in maintenance).
- [ ] /version returns appVersion, schemaVersion, goVersion, buildDate.

## Admin Channel (portable, enforced local-only)
[ ] Admin HTTP API on separate listener (--admin-port, default 8383).
[ ] Always bind to loopback only (127.0.0.1, ::1); refuse non-loopback binds (hard error).
[ ] JSON-only: reject non-application/json Content-Type/Accept.
[ ] No tokens, no remote admin, no rate-limits in v0.1.
[ ] Distinct admin mux; no admin routes on public mux.

## Frontend Rendering Strategy (v0.1)
[ ] HTMX-based public UI: server returns HTML fragments for swaps.
[ ] Templates under /templates (install, login, dashboard, etc.).
[ ] Use missing.style (formerly missing.css), AlpineJS, HTMX.
[ ] Verify session role & CSRF for all state-changing routes.

## Installation / Maintenance
[ ] Installation app (stub): serve a simple static page, “Installation is in progress.”
[ ] Maintenance mode (marker + restart):
- [ ] CLI app server maintenance on|off → /admin/maintenance/on|off writes/removes marker file in store dir.
- [ ] Require app server restart to apply.
- [ ] On startup with marker: serve installation/maintenance app; admin API stays available.
- [ ] In maintenance: public API 503 JSON or maintenance page; /ready not ready; /live OK; /admin/status shows mode: maintenance.

## Admin Commands — v0.1-alpha
### DB
[ ] app db create — Create & initialize datastore. Only command that does not use server routes.
[ ] app db upgrade — Apply migrations via /admin/db/upgrade (create timestamped backup in ./backups/).
[ ] app db verify — Read-only integrity check via /admin/db/verify.

### Server
[ ] app server status — /admin/status (version, uptime, dbVersion, mode).
[ ] app server restart — /admin/restart graceful restart (optional --delay reserved).
[ ] app server shutdown — /admin/shutdown graceful stop.
[ ] app server echo <text> — /admin/echo → { "echo": "<text>" }.

## Session Manager (contract & backends)
[ ] Opaque, random server-side sessions (no JWT).
[ ] Contract SessionStore with: Create, Get, Touch, Destroy, DestroyUserSessions, PruneExpired, Stats.
[ ] Session fields: ID, UserID, Roles[], IssuedAt, LastSeen, IdleTTL, AbsTTL, CSRF, Meta.
[ ] Constructor accepts clock and rng (deterministic tests).

### Backends
[ ] Memory: map+RWMutex, GC ticker, capacity, log evictions.
[ ] SQLite: sessions table; indexes on user_id, last_seen; JSON for roles/meta; prune by idle & absolute TTL.

### Cookies & Middleware
[ ] Cookie goob_sess: HttpOnly, Secure (TLS or X-Forwarded-Proto:https), SameSite=Lax, Path=/.
[ ] Public API routes: enforce JSON-only when applicable; HTML routes serve templates for HTMX.
[ ] Role guard helpers (RequireRole("admin")) return JSON 403 on admin API or appropriate HTML response for public routes.

### CSRF
[ ] Require maintained CSRF middleware package (pluggable CSRFMiddleware contract).
[ ] Protect all state-changing routes (public + admin).
[ ] Provide helper to expose CSRF token for HTMX/Alpine (script injection or /api/auth/csrf endpoint).
[ ] Rotate token on login.

## Configuration Model
[ ] Persisted config table app_config(key, val, type, updated_at).
[ ] Load order: defaults → persisted → flags/env at startup → one-time overrides via db create.
[ ] Bootstrap overrides (subset): --session-idle, --session-abs, --session-cookie-name, --csrf-cookie-name.
[ ] Post-boot updates via admin JSON API:
- [ ] app db update sessions session-idle 20m → /admin/config/update.
- [ ] app db show → /admin/config/list (effective + persisted).
[ ] Do not allow runtime change of admin addr/port in v0.1 (requires restart).

## Migrations & Backups
[ ] schema_migrations(version TEXT PRIMARY KEY, applied_at INTEGER); central app_version/schema_version.
[ ] Before upgrade, copy DB to ./backups/ts-name.sqlite3 and log the path.

## Error & JSON Contract
[ ] Uniform error shape: { "error": "code", "message": "human text" } with stable codes (unauthorized, forbidden, not_ready, maintenance, etc.).
[ ] JSON-only guard returns 415 with the shape above (admin and any JSON route).
[ ] Public HTML routes return proper error pages where applicable.

## Logging & Observability
[ ] Goob logging contract; default to Go std logger.
[ ] Log startup/shutdown, store checks, version mismatches, admin binds, maintenance toggles, admin command invocations.
[ ] (Deferred) Metrics hooks: active sessions, pruned count, login/logout counters.

## Store Lifecycle / Boot Behavior (recap)
[ ] Store path defaults to CWD for v0.1-alpha.
[ ] Serve installation app if store mismatch/uninitialized.
[ ] Public readiness never reveals admin mode.

## Templating hooks (for go tool template)
[ ] Template fields: {{.Module}}, {{.AppName}}, {{.BinaryName}}, {{.Port}}, {{.AdminPort}}, {{.PkgPrefix}}.
[ ] Skeleton layout:
- [ ] cmd/{{.BinaryName}}/main.go
- [ ] internal/server/*
- [ ] internal/admin/*
- [ ] internal/session/*
- [ ] internal/store/*
- [ ] templates/* (install, login, dashboard, partials)
- [ ] SECURITY_CONSIDERATIONS.md
- [ ] README.md

## Documentation
[ ] SECURITY_CONSIDERATIONS.md: enforced loopback-only admin, JSON-only admin policy, install/maintenance behavior, SQLite pragmas, deferred hardening (UDS/pipes, rate limits, mTLS), misconfiguration blocks.
[ ] Quickstart: app db create → app serve → admin CLI (status, echo, maintenance, restart).
[ ] Note on cookies & TLS behind reverse proxy; warn if Secure cookies aren’t in effect.

---
Next: v0.2 will focus on front-end UX (login page, etc.).
