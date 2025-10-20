package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultDBFile = "goobtool.db"
)

// CheckExists verifies if the datastore exists at the given path.
// Returns true if the store exists, false otherwise.
func CheckExists(storePath string) (bool, error) {
	dbPath := filepath.Join(storePath, DefaultDBFile)
	info, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check store existence: %w", err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("datastore path is a directory, expected file: %s", dbPath)
	}
	return true, nil
}

// GetStorePath returns the path to the datastore directory.
// For v0.1-alpha, this defaults to the current working directory.
func GetStorePath() string {
	return "."
}

// GetDBPath returns the full path to the database file.
func GetDBPath(storePath string) string {
	return filepath.Join(storePath, DefaultDBFile)
}
