
package logging

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// setup is a helper function to reset the logger for each test.
func setup(t *testing.T) {
	t.Helper()
	ForTestsOnlyResetLogger()
}

func TestGetLogger_DefaultInitialization(t *testing.T) {
	setup(t)

	// Capture the initial stderr.
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()

	// Restore stderr.
	w.Close()
	os.Stderr = oldStderr

	if !logger.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Default logger should have Info level enabled")
	}
	if logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Default logger should not have Debug level enabled")
	}
}

func TestInit_FirstTime(t *testing.T) {
	setup(t)

	var buf bytes.Buffer
	Init(slog.LevelDebug, &buf)

	logger := GetLogger()
	logger.Debug("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Error("Log message was not written to the buffer")
	}
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Logger should have Debug level enabled")
	}
}

func TestInit_IsNoOpAfterFirstCall(t *testing.T) {
	setup(t)

	var buf1, buf2 bytes.Buffer
	Init(slog.LevelDebug, &buf1)
	Init(slog.LevelInfo, &buf2)

	logger := GetLogger()
	logger.Debug("test message")

	if !strings.Contains(buf1.String(), "test message") {
		t.Error("Log message was not written to the first buffer")
	}
	if len(buf2.String()) > 0 {
		t.Error("Second Init call should be a no-op and not write to the second buffer")
	}
}

func TestGetLogger_ReturnsSingleton(t *testing.T) {
	setup(t)

	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger should always return the same instance")
	}

	var buf bytes.Buffer
	Init(slog.LevelDebug, &buf)
	logger3 := GetLogger()

	if logger1 != logger3 {
		t.Error("GetLogger should return the same instance even after Init")
	}
}
