package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundleGC(t *testing.T) {
	// Setup temp dir for testing
	tmpDir, err := os.MkdirTemp("", "gc-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 1. Create orphaned directory
	orphanedDir := filepath.Join(tmpDir, "orphaned-service")
	err = os.Mkdir(orphanedDir, 0750)
	assert.NoError(t, err)

	// 2. Create active directory and track it
	activeID := "active-service"
	activeDir := filepath.Join(tmpDir, activeID)
	err = os.Mkdir(activeDir, 0750)
	assert.NoError(t, err)
	trackBundle(activeID)
	defer untrackBundle(activeID)

	// 3. Run GC directly (avoiding global variable overrides and async races)
	runGC(tmpDir)

	// 4. Verify results
	_, err = os.Stat(orphanedDir)
	assert.True(t, os.IsNotExist(err), "Orphaned directory should be removed")

	_, err = os.Stat(activeDir)
	assert.NoError(t, err, "Active directory should remain")
}
