// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	// Create a temporary directory and file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, []byte("original content"), 0644)
	require.NoError(t, err)

	watcher, err := NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	err = watcher.Add(configFile)
	require.NoError(t, err)

	changeCh := make(chan struct{}, 10)
	onChange := func() error {
		changeCh <- struct{}{}
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher.Start(ctx, 50*time.Millisecond, onChange)

	// Modify the file (Write)
	time.Sleep(100 * time.Millisecond) // Give watcher time to start
	err = os.WriteFile(configFile, []byte("new content"), 0644)
	require.NoError(t, err)

	// Wait for change
	select {
	case <-changeCh:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for config change (write)")
	}

	// Atomic Save Simulation (Create/Rename)
	// Rename old file
	tmpFile := filepath.Join(tmpDir, "config.yaml.tmp")
	err = os.WriteFile(tmpFile, []byte("atomic content"), 0644)
	require.NoError(t, err)

	// Rename tmp to config (Atomic replacement)
	err = os.Rename(tmpFile, configFile)
	require.NoError(t, err)

	// Wait for change
	select {
	case <-changeCh:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for config change (atomic save)")
	}
}
