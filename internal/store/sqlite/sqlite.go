package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/maloquacious/goobtool/internal/store"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements the Store interface using modernc.org/sqlite.
type SQLiteStore struct {
	dbPath         string
	db             *sql.DB
	expectedSchema string
}

// New creates a new SQLiteStore.
func New(dbPath string, expectedSchema string) *SQLiteStore {
	return &SQLiteStore{
		dbPath:         dbPath,
		expectedSchema: expectedSchema,
	}
}

// Open opens the SQLite database with safe defaults.
func (s *SQLiteStore) Open() error {
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Apply safe defaults
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return fmt.Errorf("failed to set pragma %q: %w", pragma, err)
		}
	}

	s.db = db
	return nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// InitSchema creates the initial schema with the schema_migrations table.
func (s *SQLiteStore) InitSchema(version string) error {
	if s.db == nil {
		return fmt.Errorf("database not opened")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create schema_migrations table
	_, err = tx.Exec(initialSchema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Insert initial version
	_, err = tx.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, strftime('%s', 'now'))`, version)
	if err != nil {
		return fmt.Errorf("failed to insert schema version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CheckState returns the current state of the datastore.
func (s *SQLiteStore) CheckState() (store.StoreState, error) {
	if s.db == nil {
		return store.StateMissing, fmt.Errorf("database not opened")
	}

	// Check if schema_migrations table exists
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&count)
	if err != nil {
		return store.StateUninitialized, fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	if count == 0 {
		return store.StateUninitialized, nil
	}

	// Check schema version
	version, err := s.GetSchemaVersion()
	if err != nil {
		return store.StateUninitialized, fmt.Errorf("failed to get schema version: %w", err)
	}

	if version != s.expectedSchema {
		return store.StateVersionMismatch, nil
	}

	return store.StateReady, nil
}

// GetSchemaVersion returns the current schema version from the database.
func (s *SQLiteStore) GetSchemaVersion() (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not opened")
	}

	var version string
	err := s.db.QueryRow(`SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1`).Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to query schema version: %w", err)
	}

	return version, nil
}
