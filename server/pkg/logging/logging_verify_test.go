// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInitAddSourceBehavior(t *testing.T) {
	// Reset logger
	ForTestsOnlyResetLogger()

	var buf bytes.Buffer
	Init(slog.LevelInfo, &buf)

	l := GetLogger()
	l.Info("test message")

	if strings.Contains(buf.String(), "source=") {
		t.Errorf("Expected no source info at INFO level, but found it: %s", buf.String())
	}

	// Reset again for Debug
	ForTestsOnlyResetLogger()
	buf.Reset()
	Init(slog.LevelDebug, &buf)

	l = GetLogger()
	l.Debug("debug message")

	// Note: output format depends on text/json. Default is text.
	// TextHandler with AddSource emits "source=..."
	if !strings.Contains(buf.String(), "source=") {
		t.Errorf("Expected source info at DEBUG level, but did not find it: %s", buf.String())
	}
}
