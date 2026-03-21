// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileAuditStore_File(t *testing.T) {
	tmpDir := t.TempDir()
	validation.SetAllowedPaths([]string{tmpDir})
	defer validation.SetAllowedPaths(nil)
	logFile := filepath.Join(tmpDir, "audit.log")

	store, err := NewFileAuditStore(logFile)
	require.NoError(t, err)
	defer store.Close()

	assert.NotNil(t, store.file)
	assert.Equal(t, logFile, store.file.Name())
}

func TestNewFileAuditStore_Stdout(t *testing.T) {
	store, err := NewFileAuditStore("")
	require.NoError(t, err)
	defer store.Close()

	assert.Nil(t, store.file)
}

func TestNewFileAuditStore_Error(t *testing.T) {
	// Try to open a file in a non-existent directory to trigger error
	_, err := NewFileAuditStore("/non/existent/dir/audit.log")
	require.Error(t, err)
	// Can be either "failed to open" or "path not allowed" depending on validation
	// Since /non/existent/dir is likely not allowed, it should fail validation first
	// But let's check for Error generally, or update expectation
	// Update: it will fail validation.IsAllowedPath first.
	assert.Contains(t, err.Error(), "path not allowed")
}

func TestFileAuditStore_Write_File(t *testing.T) {
	tmpDir := t.TempDir()
	validation.SetAllowedPaths([]string{tmpDir})
	defer validation.SetAllowedPaths(nil)
	logFile := filepath.Join(tmpDir, "audit.log")

	store, err := NewFileAuditStore(logFile)
	require.NoError(t, err)
	defer store.Close()

	entry := Entry{
		ToolName: "test-tool",
		Error:    "test-error", // Replacing Status with Error, as Status doesn't exist
	}

	err = store.Write(context.Background(), entry)
	require.NoError(t, err)

	// Read file content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var readEntry Entry
	err = json.Unmarshal(content, &readEntry)
	require.NoError(t, err)
	assert.Equal(t, entry.ToolName, readEntry.ToolName)
	assert.Equal(t, entry.Error, readEntry.Error)
}

func TestFileAuditStore_Write_Stdout(t *testing.T) {
	store, err := NewFileAuditStore("")
	require.NoError(t, err)
	defer store.Close()

	// Inject buffer instead of os.Stdout to avoid race conditions
	var buf bytes.Buffer
	store.out = &buf

	entry := Entry{
		ToolName: "stdout-tool",
		Error:    "failure",
	}

	err = store.Write(context.Background(), entry)
	require.NoError(t, err)

	var readEntry Entry
	err = json.Unmarshal(buf.Bytes(), &readEntry)
	require.NoError(t, err)
	assert.Equal(t, entry.ToolName, readEntry.ToolName)
	assert.Equal(t, entry.Error, readEntry.Error)
}

func TestFileAuditStore_Close(t *testing.T) {
	tmpDir := t.TempDir()
	validation.SetAllowedPaths([]string{tmpDir})
	defer validation.SetAllowedPaths(nil)
	logFile := filepath.Join(tmpDir, "audit.log")

	store, err := NewFileAuditStore(logFile)
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)

	// Closing again should be fine (file.Close is idempotent usually, or returns error if already closed)
	// In os.File, Close returns error if already closed.
	err = store.Close()
	assert.Error(t, err) // Expect error on second close if implementation just calls file.Close()
}

func TestFileAuditStore_Close_Stdout(t *testing.T) {
	store, err := NewFileAuditStore("")
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)
}

func TestFileAuditStore_Read_Error(t *testing.T) {
	store, err := NewFileAuditStore("")
	require.NoError(t, err)
	defer store.Close()

	entries, err := store.Read(context.Background(), Filter{})
	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "read not implemented")
}
