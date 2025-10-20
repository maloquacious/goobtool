# Sprint 3 Implementation Summary

## Completed Tasks

Sprint 3 has been successfully implemented with all requirements met:

### ✅ Core Requirements
- **SQLite datastore** via modernc.org/sqlite (portable, no CGo)
- **Store interface** (Goob contract) for swappable backends
- **SQLite safe defaults**: WAL mode, synchronous=NORMAL, foreign_keys=ON, busy_timeout=5000
- **app db create command** - Creates and initializes datastore with schema_migrations table
- **Enhanced lifecycle checks** - Detects uninitialized, version mismatch, and ready states
- **Installation page** - Served when datastore needs attention
- **Health endpoints**: /live (always OK), /ready (OK when initialized)
- **/version endpoint** - Returns appVersion, schemaVersion, goVersion, buildDate
- **Proper routing** - Serves public/index.html when store is ready

## Implementation Details

### New Packages

#### `internal/store/interface.go`
- `Store` interface defining the Goob datastore contract
- `StoreState` enum: Missing, Uninitialized, VersionMismatch, Ready
- Methods: Open, Close, InitSchema, CheckState, GetSchemaVersion

#### `internal/store/sqlite/`
- `sqlite.go` - SQLiteStore implementation of Store interface
- `schema.go` - Initial schema with schema_migrations table
- Safe defaults applied via PRAGMAs on Open()

### SQLite Configuration

```go
PRAGMA journal_mode=WAL       // Write-Ahead Logging for better concurrency
PRAGMA synchronous=NORMAL     // Balance between safety and performance
PRAGMA foreign_keys=ON        // Enforce referential integrity
PRAGMA busy_timeout=5000      // 5 second timeout for locked database
```

### Datastore States

1. **Missing** - Database file doesn't exist → Exit with guidance
2. **Uninitialized** - File exists but no schema_migrations table → Serve installation page
3. **VersionMismatch** - Schema version doesn't match expected → Serve installation page  
4. **Ready** - Initialized and correct version → Serve normal application

### Health & Version Endpoints

#### Normal Operation (Store Ready)
- **GET /live** → 200 OK (always)
- **GET /ready** → 200 OK (store initialized)
- **GET /version** → JSON with versions

#### Installation Mode (Store Needs Attention)
- **GET /live** → 200 OK (process is running)
- **GET /ready** → 503 Service Unavailable (not ready)
- **GET /version** → JSON with versions
- **GET /** → install.html page

### Database Schema

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at INTEGER NOT NULL  -- Unix timestamp
);
```

### Command Implementation

#### `app db create`
```bash
$ app db create

✓ Datastore created successfully
  Path: goobtool.db
  Schema version: 0.1
```

- Checks if database already exists (fails if it does)
- Creates database file
- Opens with safe PRAGMA settings
- Creates schema_migrations table
- Inserts initial version record
- Cleans up on failure

## Testing

### Manual Tests

```bash
# Test without database
./dist/local/app serve
# → Exits with "Run: app db create"

# Create database
./dist/local/app db create
# → Creates goobtool.db with schema 0.1

# Test with initialized database
./dist/local/app serve --exit-after=2s
# → Starts normally, serves application

# Test uninitialized database (drop migrations table)
sqlite3 goobtool.db "DROP TABLE schema_migrations;"
./dist/local/app serve --exit-after=2s
# → Serves installation page

# Test health endpoints
curl http://localhost:8080/live      # → 200 OK
curl http://localhost:8080/ready     # → 200 OK (or 503 in installation mode)
curl http://localhost:8080/version   # → JSON with version info
```

### Automated Test Script
- ✅ `tmp/test_sprint3.sh` - Tests health endpoints, version, and database schema

All tests passing.

## Log Output Examples

### Successful startup with initialized database:
```
[INFO] starting Goobergine server version=0.1.3-alpha schema=0.1
[INFO] datastore ready path=goobtool.db schema=0.1
[INFO] admin listener verified on loopback: 127.0.0.1:8383
[INFO] public server listening on port=8080
[INFO] admin server listening on 127.0.0.1:8383 (JSON-only)
```

### Uninitialized database (installation mode):
```
[INFO] starting Goobergine server version=0.1.3-alpha schema=0.1
[WARN] datastore uninitialized (missing schema_migrations table)
[INFO] serving installation app (datastore requires attention)
[INFO] public server listening on port=8080 (installation mode)
```

### Database creation:
```
[INFO] creating datastore schema=0.1
[INFO] datastore created successfully path=goobtool.db schema=0.1
```

## Next Steps

Sprint 3 is complete and ready for Sprint 4, which will implement:
- Maintenance mode (marker + restart)
- app db verify command
- app server status/shutdown/echo commands
- Store path defaults
- Error contract enforcement
