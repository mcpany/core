package audit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLiteAuditStore_ErrorCases(t *testing.T) {
	t.Run("empty_path", func(t *testing.T) {
		store, err := NewSQLiteAuditStore("")
		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "sqlite path is required")
	})

	t.Run("invalid_path_directory_not_exist", func(t *testing.T) {
		// Attempt to open a DB in a non-existent directory
		store, err := NewSQLiteAuditStore("/non/existent/directory/audit.db")
		assert.Error(t, err)
		assert.Nil(t, store)
		// Error message might depend on OS/Driver, but it should fail
	})

	t.Run("invalid_permission", func(t *testing.T) {
		// Create a directory with read-only permissions
		tempDir := t.TempDir()
		// No write permission
		// Note: t.TempDir() creates a dir with 0700 usually.
		// We can try to use a path that is a directory instead of a file
		_ = filepath.Join(tempDir, "audit.db")
		// make the directory read only ?
		// easier: point to a directory as if it was a file
		store, err := NewSQLiteAuditStore(tempDir) // tempDir is a directory, not a file
		assert.Error(t, err)
		assert.Nil(t, store)
	})
}
