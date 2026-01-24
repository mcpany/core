// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogWriter(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	writer := NewLogWriter(log)

	testMessage := "Hello, world!\n"
	_, err := writer.Write([]byte(testMessage))
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), "Hello, world!")
	assert.Contains(t, buf.String(), "INFO")
}

func TestLogWriter_LevelDetection(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	log := slog.New(slog.NewJSONHandler(&buf, opts))
	writer := NewLogWriter(log)

	// Test Error detection
	writer.Write([]byte("fatal error occurred\n"))
	assert.Contains(t, buf.String(), "ERROR")
	assert.Contains(t, buf.String(), "fatal error occurred")
	buf.Reset()

	// Test Warn detection
	writer.Write([]byte("warning: check config\n"))
	assert.Contains(t, buf.String(), "WARN")
	buf.Reset()

	// Test explicit prefix
	writer.Write([]byte("[DEBUG] something low level\n"))
	assert.Contains(t, buf.String(), "DEBUG")
	buf.Reset()

	// Test buffering (partial writes)
	writer.Write([]byte("part1 "))
	assert.Empty(t, buf.String())
	writer.Write([]byte("part2\n"))
	assert.Contains(t, buf.String(), "part1 part2")
}
