package store

// StoreState represents the initialization state of the datastore.
type StoreState int

const (
	StateMissing         StoreState = iota // File doesn't exist
	StateUninitialized                     // File exists but no schema
	StateVersionMismatch                   // Schema exists but wrong version
	StateReady                             // Initialized and correct version
)

// Store defines the Goob datastore contract.
// Implementations must be safe for concurrent use.
type Store interface {
	// Open opens the datastore connection
	Open() error

	// Close closes the datastore connection
	Close() error

	// InitSchema creates the initial schema (schema_migrations table)
	InitSchema(version string) error

	// CheckState returns the current state of the datastore
	CheckState() (StoreState, error)

	// GetSchemaVersion returns the current schema version from the database
	GetSchemaVersion() (string, error)
}
