# Goobergine v0.1-alpha Summary

## 🎯 Goal
Establish a self-contained Go application generator with:
- Go backend using **standard library first** principles.
- HTMX/AlpineJS frontend (HTML fragments only).
- Local-only JSON admin API.
- SQLite datastore via `modernc.org/sqlite`.
- Fully portable, single-binary build for Windows, macOS, and Linux.

---

## 📂 Deliverables Created

| File | Purpose |
|------|----------|
| **TODO.md** | Master checklist for v0.1 features and architecture. |
| **README.md** | Overview, quickstart, directory layout, and roadmap. |
| **SECURITY_CONSIDERATIONS.md** | Loopback-only admin port, JSON-only rules, CSRF protection, deferred v1 features. |
| **LICENSE** | MIT license for public distribution. |
| **Makefile** | Build, run, test, tidy, and clean targets. |
| **AGENT.md** | Developer agent guide — stack, build/test flow, contracts, and philosophy. |
| **cmd/app/main.go** | Runnable entry point with Cobra CLI, public/admin HTTP servers, and stubbed functions (`// TODO` markers). |
| **public/index.html** | Minimal landing page with Missing.css, AlpineJS, and HTMX (via CDN). |

---

## ⚙️ Architecture Highlights

- **Single binary** includes both web app and admin CLI.  
- **Public server:** serves HTML fragments via HTMX.  
- **Admin server:** JSON-only, loopback-only (127.0.0.1).  
- **Graceful shutdown** and `--exit-after` timer for testing.  
- **Cobra CLI:**  
  ```
  app serve --port 8080 --admin-port 8383
  app db create | upgrade | verify
  app server restart | shutdown | status | echo
  ```
- **SQLite store:** portable, with WAL mode and migration hooks.  
- **Configuration:** persisted in `app_config`; runtime flags override defaults.  
- **Session manager:** interface-driven, with in-memory + SQLite backends planned.  
- **Security:** CSRF middleware required; cookies set as `Secure`, `HttpOnly`, `SameSite=Lax`.

---

## 🧩 Directories

```
cmd/app/          # CLI + server entry point
internal/server/  # Public routes (HTML)
internal/admin/   # Admin JSON API
internal/session/ # Session interfaces/backends
internal/store/   # SQLite schema/migrations
public/           # Static + templates
```

---

## 🧰 Build & Test Flow

```bash
make build           # builds dist/app
make run             # runs both ports (8080 / 8383)
make test            # runs all tests
make tidy            # go mod tidy
```

Unit tests will be added for `SessionStore` and contract compliance in v0.2.  
Integration tests will verify admin routes and HTML fragment responses.

---

## 🔐 Security Recap

- Admin API: **loopback-only**, **JSON-only**, **no auth in v0.1**.  
- Public API: CSRF-protected HTML routes only.  
- CORS disabled by default.  
- Secure cookies enforced when proxied via HTTPS.  
- Maintenance mode toggled via CLI; served as a simple static page.  

---

## 🧭 Roadmap

| Version | Focus |
|----------|--------|
| **v0.1-alpha** | Core architecture, CLI, dual servers, scaffolding |
| **v0.2** | Frontend UX (login, session UI, templates, CSRF injection) |
| **v0.3** | Admin configuration editor |
| **v0.4** | RBAC, richer logging, runtime config |
| **v1.0** | Remote admin, full test suite, mTLS, UDS support |

---

_This summary marks the completion of Goobergine v0.1-alpha setup and the foundation for the v0.2 frontend milestone._
