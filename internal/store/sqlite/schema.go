package sqlite

// initialSchema is the minimal schema for v0.1-alpha.
// Contains only the schema_migrations table for version tracking.
const initialSchema = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at INTEGER NOT NULL
);
`
