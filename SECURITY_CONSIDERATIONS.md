# Goobergine Application Generator — Security Considerations (v0.1-alpha)

This document explains the **security posture** for Goobergine v0.1-alpha and identifies items deferred for later releases.

## 1. Admin Channel

- The **admin API** runs on a **dedicated listener** (`--admin-port`, default `8383`).
- It **always binds to loopback** (`127.0.0.1`, `::1`) and **refuses non-loopback binds**.
- **Remote admin** is not supported in v0.1.
- All admin routes are **JSON-only** and require `Content-Type: application/json` and `Accept: application/json`.
- No tokens, no authentication headers, no rate limits (deferred to v0.2+).
- Any misconfiguration that attempts to bind the admin listener to a public interface results in a **hard error**.

## 2. Public Web Application

- The public server (`--port`, default `8080`) serves HTML fragments for HTMX clients.
- All user-facing routes verify sessions and CSRF protection via a maintained Go CSRF middleware package.
- Cookies use secure defaults: `HttpOnly`, `SameSite=Lax`, and `Secure` when behind TLS or proxy with `X-Forwarded-Proto: https`.
- CORS is disabled by default.

## 3. Sessions

- Server-side opaque session IDs (no JWT).
- IDs are random 32-byte CSPRNG values (base64url encoded).
- Session backends: `sqlite` (default) or in-memory (for testing).
- Session IDs are **rotated** on login and privilege elevation.
- Idle and absolute TTLs are enforced server-side; both configurable.

## 4. CSRF Protection

- CSRF protection is enforced by a maintained third-party middleware.
- All state-changing requests require a valid token and matching cookie.
- Tokens are rotated on login.
- For HTMX, a helper exposes tokens via script injection or `/api/auth/csrf` endpoint.

## 5. Maintenance & Installation Modes

- Maintenance is controlled by a **marker file** in the store directory.
- Commands: `app server maintenance on|off` toggle maintenance.
- When active, public routes return 503 or serve the maintenance page.
- Admin listener remains available on loopback for upgrades or shutdown.

## 6. SQLite Database Safety

- Default pragmas: `WAL`, `synchronous=NORMAL`, `foreign_keys=ON`.
- Backups are created before migrations (`./backups/` directory).
- The schema includes `schema_migrations` for version tracking and `app_config` for runtime settings.

## 7. Error Contract

- All JSON API errors follow: `{ "error": "code", "message": "human text" }`.
- Common codes: `unauthorized`, `forbidden`, `not_ready`, `maintenance`, `invalid_request`.
- Non-JSON requests to JSON-only endpoints return `415` with an error object.

## 8. Logging & Observability

- Logs use the Goob logging contract; default is Go’s std logger.
- Sensitive values (cookies, CSRF tokens, passwords) are **never logged**.
- Admin binds, shutdowns, and migration results are logged at INFO.

## 9. Deferred Until v1

- Remote/mTLS-protected admin access.
- Unix Domain Sockets (Linux/macOS) and Named Pipes (Windows).
- Rate limiting and abuse detection.
- Audit logging and immutable logs.
- Configurable CORS.
- Secret rotation policies.
- Security headers (CSP, HSTS, etc.).

## 10. Recommended User Actions

1. **Restrict firewall rules** to prevent exposing `--admin-port` externally.
2. **Serve behind a TLS-enabled reverse proxy** (Caddy, Nginx, etc.).
3. **Back up the database** before running upgrades or migrations.
4. **Rotate cookies** if session configuration changes.
5. **Test maintenance mode** before live deployments.

---
Future versions will introduce stronger access control, optional remote administration, and deeper runtime hardening.
