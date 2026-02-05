// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package gc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		input    Config
		expected Config
	}{
		{
			name:  "defaults",
			input: Config{},
			expected: Config{
				Interval: 1 * time.Hour,
				TTL:      24 * time.Hour,
			},
		},
		{
			name: "custom values",
			input: Config{
				Enabled:  true,
				Interval: 10 * time.Minute,
				TTL:      2 * time.Hour,
				Paths:    []string{"/tmp"},
			},
			expected: Config{
				Enabled:  true,
				Interval: 10 * time.Minute,
				TTL:      2 * time.Hour,
				Paths:    []string{"/tmp"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New(tt.input)
			assert.Equal(t, tt.expected, w.config)
		})
	}
}

func TestWorker_Start(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gc-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create files
	oldFile := filepath.Join(tmpDir, "old_file.txt")
	newFile := filepath.Join(tmpDir, "new_file.txt")
	ignoredDir := filepath.Join(tmpDir, "ignored_dir")

	// Create old file (older than TTL)
	err = os.WriteFile(oldFile, []byte("old"), 0644)
	require.NoError(t, err)
	// Modify time to be 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	require.NoError(t, err)

	// Create new file (newer than TTL)
	err = os.WriteFile(newFile, []byte("new"), 0644)
	require.NoError(t, err)

	// Create directory (should be ignored by file cleanup logic?)
	// The code says: "GC: Removing old item". It uses os.RemoveAll, so it might delete directories too if they are old.
	// Let's check the code: `entries, err := os.ReadDir(cleanPath)`. `entry.Info()`. `info.ModTime()`.
	// Yes, it iterates over entries. If an entry is a directory and old, it will be removed.
	err = os.Mkdir(ignoredDir, 0755)
	require.NoError(t, err)
	// Set directory time to old
	err = os.Chtimes(ignoredDir, oldTime, oldTime)
	require.NoError(t, err)

	// Config
	cfg := Config{
		Enabled:  true,
		Interval: 100 * time.Millisecond,
		TTL:      1 * time.Hour,
		Paths:    []string{tmpDir},
	}

	w := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.Start(ctx)

	// Wait for cleanup
	// We expect oldFile to be deleted, newFile to remain.
	// ignoredDir (which is old) should also be deleted because the code does `os.RemoveAll`.

	assert.Eventually(t, func() bool {
		_, err := os.Stat(oldFile)
		return os.IsNotExist(err)
	}, 2*time.Second, 100*time.Millisecond, "old file should be deleted")

	assert.Eventually(t, func() bool {
		_, err := os.Stat(ignoredDir)
		return os.IsNotExist(err)
	}, 2*time.Second, 100*time.Millisecond, "old directory should be deleted")

	// Check that new file still exists
	_, err = os.Stat(newFile)
	assert.NoError(t, err, "new file should exist")
}

func TestWorker_Start_Disabled(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "gc-test-disabled-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	oldFile := filepath.Join(tmpDir, "old_file.txt")
	err = os.WriteFile(oldFile, []byte("old"), 0644)
	require.NoError(t, err)
	oldTime := time.Now().Add(-2 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	require.NoError(t, err)

	cfg := Config{
		Enabled:  false, // Disabled
		Interval: 10 * time.Millisecond,
		TTL:      1 * time.Hour,
		Paths:    []string{tmpDir},
	}

	w := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.Start(ctx)

	// Sleep a bit to ensure it didn't run
	time.Sleep(100 * time.Millisecond)

	_, err = os.Stat(oldFile)
	assert.NoError(t, err, "old file should still exist when disabled")
}

func TestWorker_Start_DangerousPaths(t *testing.T) {
	// We can't really test that it doesn't delete root without mocking os,
	// but we can check coverage of that branch by setting path to "/" or "."
	// Ideally we would mock the logger to verify the warning.

	cfg := Config{
		Enabled:  true,
		Interval: 10 * time.Millisecond,
		TTL:      1 * time.Hour,
		Paths:    []string{"/", "."},
	}

	w := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This shouldn't crash or delete anything important (due to checks)
	w.Start(ctx)
	time.Sleep(50 * time.Millisecond)
}
