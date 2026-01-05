// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileAuditStore_File(t *testing.T) {
	tmpDir := t.TempDir()
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
	assert.Contains(t, err.Error(), "failed to open audit log file")
}

func TestFileAuditStore_Write_File(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "audit.log")

	store, err := NewFileAuditStore(logFile)
	require.NoError(t, err)
	defer store.Close()

	entry := AuditEntry{
		ToolName: "test-tool",
		Error:    "test-error", // Replacing Status with Error, as Status doesn't exist
	}

	err = store.Write(context.Background(), entry)
	require.NoError(t, err)

	// Read file content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var readEntry AuditEntry
	err = json.Unmarshal(content, &readEntry)
	require.NoError(t, err)
	assert.Equal(t, entry.ToolName, readEntry.ToolName)
	assert.Equal(t, entry.Error, readEntry.Error)
}

func TestFileAuditStore_Write_Stdout(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	store, err := NewFileAuditStore("")
	require.NoError(t, err)
	defer store.Close()

	entry := AuditEntry{
		ToolName: "stdout-tool",
		Error:    "failure",
	}

	err = store.Write(context.Background(), entry)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	var readEntry AuditEntry
	err = json.Unmarshal(buf.Bytes(), &readEntry)
	require.NoError(t, err)
	assert.Equal(t, entry.ToolName, readEntry.ToolName)
	assert.Equal(t, entry.Error, readEntry.Error)
}

func TestFileAuditStore_Close(t *testing.T) {
	tmpDir := t.TempDir()
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
