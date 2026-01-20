// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitIgnoredAfterGetLogger(t *testing.T) {
	// Reset logger state for test
	ForTestsOnlyResetLogger()

	// 1. Call GetLogger first. This initializes with defaults (Info level).
	log1 := GetLogger()
	assert.NotNil(t, log1)
	// Default is Info. Debug should be disabled.
	assert.False(t, log1.Enabled(context.Background(), slog.LevelDebug), "Default logger should not log Debug")

	// 2. Call Init with Debug level.
	var buf bytes.Buffer
	Init(slog.LevelDebug, &buf, "text")

	// 3. GetLogger again.
	log2 := GetLogger()

	// 4. Verify log2 respects Init settings.
	// If bug exists, log2 is still the default logger (Info only).
	// Current behavior: Init does nothing because GetLogger used sync.Once.
	if !log2.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("Logger should be enabled for Debug after Init")
	}

	// Also verify output destination changed
	log2.Debug("test debug")
	if buf.Len() == 0 {
		t.Error("Logger should write to buffer after Init")
	}
}
