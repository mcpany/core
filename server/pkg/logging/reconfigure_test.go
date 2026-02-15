// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconfigure_Level(t *testing.T) {
	// Reset
	ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	Init(slog.LevelInfo, &buf, "")

	l := GetLogger()
	l.Debug("debug message 1") // Should be ignored
	l.Info("info message 1")

	assert.NotContains(t, buf.String(), "debug message 1")
	assert.Contains(t, buf.String(), "info message 1")

	// Reconfigure to Debug
	Reconfigure(slog.LevelDebug, "text")
	l = GetLogger()

	l.Debug("debug message 2")

	assert.Contains(t, buf.String(), "debug message 2")
}

func TestReconfigure_File(t *testing.T) {
	ForTestsOnlyResetLogger()

	tmpFile, err := os.CreateTemp("", "logtest")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close() // Close so Init can open it
	defer os.Remove(tmpPath)

	Init(slog.LevelInfo, os.Stderr, tmpPath, "json")

	l := GetLogger()
	l.Info("message to file")

	// Check file content
	// We might need to wait a bit or ensure flush, but slog is usually synchronous for file IO in basic handler
	content, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "message to file")

	// Reconfigure - should rotate/reopen file
	Reconfigure(slog.LevelInfo, "json")
	l = GetLogger()
	l.Info("message after reconfigure")

	// Check file content again (should append)
	content, err = os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "message after reconfigure")
}
