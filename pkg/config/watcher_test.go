package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	// Create a temporary directory for the test files
	tmpDir, err := os.MkdirTemp("", "watcher-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test configuration file
	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, []byte("test: initial"), 0644)
	require.NoError(t, err)

	// Create a channel to receive reload signals
	reloadChan := make(chan struct{}, 1)

	// Create a new watcher
	watcher, err := NewWatcher(reloadChan)
	require.NoError(t, err)
	defer watcher.Close()

	// Start watching the configuration file
	err = watcher.Watch([]string{configPath})
	require.NoError(t, err)

	// Modify the configuration file
	err = os.WriteFile(configPath, []byte("test: modified"), 0644)
	require.NoError(t, err)

	// Check if a reload signal is received
	select {
	case <-reloadChan:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reload signal")
	}
}
