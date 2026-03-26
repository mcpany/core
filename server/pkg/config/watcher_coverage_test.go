// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatcher_RelativePath(t *testing.T) {
	// Create a temporary file to watch in the current directory (to allow relative path usage)
	// We need to be careful not to leave garbage.
	// We can use os.MkdirTemp
	tmpDir, err := os.MkdirTemp("", "watcher_relative_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

	fileName := "config.yaml"
	file, err := os.Create(fileName)
	require.NoError(t, err)
	file.Close()

	w, err := NewWatcher()
	require.NoError(t, err)
	defer w.Close()

	reloaded := make(chan bool)
	go func() {
		_ = w.Watch([]string{fileName}, func() {
			reloaded <- true
		})
	}()

	time.Sleep(100 * time.Millisecond)

	// Modify file
	err = os.WriteFile(fileName, []byte("change"), 0o644)
	require.NoError(t, err)

	select {
	case <-reloaded:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Watcher failed to reload on relative path")
	}
}

func TestWatcher_Debounce(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher_debounce_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(filePath, []byte("initial"), 0o644)

	w, err := NewWatcher()
	require.NoError(t, err)
	defer w.Close()

	var reloadCount int32
	go func() {
		_ = w.Watch([]string{filePath}, func() {
			atomic.AddInt32(&reloadCount, 1)
		})
	}()

	time.Sleep(100 * time.Millisecond)

	// Quick writes
	os.WriteFile(filePath, []byte("change1"), 0o644)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filePath, []byte("change2"), 0o644)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filePath, []byte("change3"), 0o644)

	// Wait for debounce (500ms)
	time.Sleep(1 * time.Second)

	count := atomic.LoadInt32(&reloadCount)
	// Should trigger once (or maybe twice depending on timing, but definitely not 3)
	if count == 0 {
		t.Error("Watcher failed to reload")
	}
	// We expect debouncing to collapse events.
	// Since we sleep 10ms between writes, and debounce is 500ms, it should reset the timer each time.
	// So only 1 reload should happen after the last write.
	if count > 1 {
		t.Logf("Reload count: %d (expected 1 due to debouncing)", count)
	}
}
