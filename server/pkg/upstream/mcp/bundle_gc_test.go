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

func TestBundleGC_TriggerGCTime(t *testing.T) {
	// Setup temp dir for testing
	tmpDir, err := os.MkdirTemp("", "gc-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	bundleBaseDir = tmpDir

	// 1. Create orphaned directory
	orphanedDir := filepath.Join(tmpDir, "orphaned-service-2")
	err = os.Mkdir(orphanedDir, 0750)
	assert.NoError(t, err)

	// Set last GC to far past to force a trigger
	lastGCTimestamp.Store(0)
	triggerGC()

	// Wait a moment for background goroutine to execute with eventually
	assert.Eventually(t, func() bool {
		_, err := os.Stat(orphanedDir)
		return os.IsNotExist(err)
	}, 2*time.Second, 10*time.Millisecond, "Orphaned directory should be removed by background GC")

	// Create another orphaned and trigger again, but it shouldn't run this time because of the interval
	orphanedDir3 := filepath.Join(tmpDir, "orphaned-service-3")
	err = os.Mkdir(orphanedDir3, 0750)
	assert.NoError(t, err)

	triggerGC()

	// Ensure that even after waiting a bit, it still is not deleted (negative assertion)
	// We wait explicitly here because we want to prove it DID NOT happen within a timeframe
	time.Sleep(100 * time.Millisecond)

	_, err = os.Stat(orphanedDir3)
	assert.NoError(t, err, "Orphaned directory should NOT be removed by background GC because of interval")
}
