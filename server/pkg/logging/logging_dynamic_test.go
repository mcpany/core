// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicLevel(t *testing.T) {
	// Reset global state for test
	ForTestsOnlyResetLogger()

	var buf bytes.Buffer
	// Initialize with INFO level
	Init(slog.LevelInfo, &buf, "", "json")

	logger := GetLogger()

	// 1. Verify INFO logs show up
	logger.Info("info message 1")
	assert.Contains(t, buf.String(), "info message 1")
	buf.Reset()

	// 2. Verify DEBUG logs do NOT show up
	logger.Debug("debug message 1")
	assert.NotContains(t, buf.String(), "debug message 1")
	buf.Reset()

	// 3. Change level to DEBUG dynamically
	SetLevel(slog.LevelDebug)

	// 4. Verify DEBUG logs NOW show up
	logger.Debug("debug message 2")
	assert.Contains(t, buf.String(), "debug message 2")
	buf.Reset()

	// 5. Change level back to WARN
	SetLevel(slog.LevelWarn)

	// 6. Verify INFO logs do NOT show up
	logger.Info("info message 2")
	assert.NotContains(t, buf.String(), "info message 2")
	buf.Reset()

	// 7. Verify WARN logs DO show up
	logger.Warn("warn message 1")
	assert.Contains(t, buf.String(), "warn message 1")
}
