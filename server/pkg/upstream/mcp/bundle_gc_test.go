// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBundleGC(t *testing.T) {
	// Setup temp dir for testing
	tmpDir, err := os.MkdirTemp("", "gc-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Override bundleBaseDir
	origBaseDir := bundleBaseDir
	bundleBaseDir = tmpDir
	defer func() { bundleBaseDir = origBaseDir }()

	// Override gcInterval
	origInterval := gcInterval
	gcInterval = 10 * time.Millisecond
	defer func() { gcInterval = origInterval }()

	// Reset lastGCTimestamp
	lastGCTimestamp.Store(0)

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

	// 3. Trigger GC
	triggerGC()

	// Wait for GC (it runs in background)
	time.Sleep(100 * time.Millisecond)

	// 4. Verify results
	_, err = os.Stat(orphanedDir)
	assert.True(t, os.IsNotExist(err), "Orphaned directory should be removed")

	_, err = os.Stat(activeDir)
	assert.NoError(t, err, "Active directory should remain")
}
