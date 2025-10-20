# Sprint 2 Implementation Summary

## Completed Tasks

Sprint 2 has been successfully implemented with all requirements met:

### ✅ Core Requirements
- **Admin HTTP API on separate listener** via `--admin-port` (default: 8383)
- **Loopback-only binding** with support for both IPv4 (127.0.0.1) and IPv6 (::1)
- **Hard error on non-loopback** - Validates admin host before and after binding
- **JSON-only enforcement** - Rejects non-JSON Content-Type and Accept headers
- **Distinct admin mux** - Admin routes completely separated from public routes
- **Graceful shutdown** - Admin listener properly shut down with public server

## Implementation Details

### New Features

#### IPv4 and IPv6 Loopback Support
- Added `--admin-host` flag (default: "127.0.0.1")
- Supports both IPv4 (127.0.0.1) and IPv6 (::1) loopback addresses
- Pre-validation: Checks IP is loopback before attempting to bind
- Post-validation: Verifies bound address is loopback (defense in depth)

#### Loopback-Only Enforcement
```go
// Pre-validation
adminIP := net.ParseIP(adminHost)
if adminIP == nil || !adminIP.IsLoopback() {
    log.Error("admin host must be loopback (127.0.0.1 or ::1), got: %s", adminHost)
    os.Exit(1)
}

// Post-validation (defense in depth)
if addr, ok := adminListener.Addr().(*net.TCPAddr); ok {
    if !addr.IP.IsLoopback() {
        log.Error("admin listener bound to non-loopback address: %s", addr.IP)
        adminListener.Close()
        os.Exit(1)
    }
}
```

#### Code Cleanup
- Replaced `fmt.Sprintf` with `net.JoinHostPort` for address construction
- More idiomatic Go networking code

### Admin Endpoints

All endpoints require `Accept: application/json` header:

- **GET /admin/status** - Returns server status, version, and uptime
- **GET /admin/echo?q=text** - Simple echo endpoint for testing
- **POST /admin/echo** - JSON echo endpoint `{"echo": "text"}`
- **POST /admin/shutdown** - Graceful server shutdown
- **POST /admin/restart** - Server restart (exit 0)

### JSON-Only Enforcement

The `jsonOnly` middleware enforces:
- **406 Not Acceptable** - If Accept header doesn't include application/json
- **415 Unsupported Media Type** - If POST/PUT without Content-Type: application/json

## Testing

### Manual Tests
```bash
# Build
go build -o dist/local/app ./cmd/app

# Test IPv4 loopback (default)
cd tmp && ../dist/local/app serve

# Test IPv6 loopback
cd tmp && ../dist/local/app serve --admin-host=::1

# Test non-loopback rejection
cd tmp && ../dist/local/app serve --admin-host=0.0.0.0
# Should exit with error
```

### Automated Test Scripts
- ✅ `tmp/test_admin.sh` - Tests all IPv4 admin endpoints and JSON enforcement
- ✅ `tmp/test_ipv6.sh` - Tests IPv6 admin endpoints

All tests pass successfully.

## Security Features

1. **Loopback-only binding** - Admin API only accessible from localhost
2. **Dual validation** - Pre-bind and post-bind loopback verification
3. **JSON-only API** - Prevents accidental HTML/browser access
4. **No authentication in v0.1** - Documented limitation, deferred to v1.0
5. **Separate mux** - Admin routes isolated from public application

## Log Output Examples

### Successful IPv4 startup:
```
[INFO] starting Goobergine server version=0.1.1-alpha schema=0.1
[INFO] datastore found at path=.
[INFO] admin listener verified on loopback: 127.0.0.1:8383
[INFO] admin server listening on 127.0.0.1:8383 (JSON-only)
[INFO] public server listening on port=8080
```

### Successful IPv6 startup:
```
[INFO] admin listener verified on loopback: [::1]:8383
[INFO] admin server listening on 127.0.0.1:8383 (JSON-only)
```

### Non-loopback rejection:
```
[ERROR] admin host must be loopback (127.0.0.1 or ::1), got: 0.0.0.0
```

## Next Steps

Sprint 2 is complete and ready for Sprint 3, which will implement:
- Datastore creation (SQLite with modernc.org/sqlite)
- Storage interfaces (Goob contracts)
- Safe defaults (WAL, synchronous=NORMAL, foreign_keys=ON)
- Basic schema versioning/migration table
- Health endpoints (/live, /ready)
- Version endpoint
