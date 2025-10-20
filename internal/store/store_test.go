package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func(string) error
		wantExist bool
		wantError bool
	}{
		{
			name: "database exists",
			setup: func(dir string) error {
				dbPath := filepath.Join(dir, DefaultDBFile)
				f, err := os.Create(dbPath)
				if err != nil {
					return err
				}
				return f.Close()
			},
			wantExist: true,
			wantError: false,
		},
		{
			name: "database does not exist",
			setup: func(dir string) error {
				return nil
			},
			wantExist: false,
			wantError: false,
		},
		{
			name: "database path is directory",
			setup: func(dir string) error {
				dbPath := filepath.Join(dir, DefaultDBFile)
				return os.Mkdir(dbPath, 0755)
			},
			wantExist: false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.Mkdir(testDir, 0755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			if err := tt.setup(testDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			exists, err := CheckExists(testDir)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if exists != tt.wantExist {
				t.Errorf("got exists=%v, want %v", exists, tt.wantExist)
			}
		})
	}
}

func TestGetStorePath(t *testing.T) {
	path := GetStorePath()
	if path != "." {
		t.Errorf("got %q, want %q", path, ".")
	}
}
