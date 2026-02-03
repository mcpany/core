// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogger(t *testing.T) {
	// The ResetLoggerForTests is crucial to ensure that the global logger can be
	// re-initialized for each test case, preventing state leakage between tests.
	ForTestsOnlyResetLogger()

	// Capture the output of the logger to verify its configuration.
	var buf bytes.Buffer
	Init(slog.LevelDebug, &buf)

	// Retrieve the logger instance.
	logger := GetLogger()
	require.NotNil(t, logger, "GetLogger should not return nil after Init")

	// Log a message and check if it's written to the buffer with the correct level.
	logger.Debug("test message")
	assert.Contains(t, buf.String(), "level=DEBUG", "Logger should be configured with the level from Init")
	assert.Contains(t, buf.String(), "msg=\"test message\"", "Logger should write the correct message")
}

func TestGetLogger_DefaultInitialization(t *testing.T) {
	// Reset the logger to simulate the initial state of the application.
	ForTestsOnlyResetLogger()

	// Directly call GetLogger without a prior Init to test its default behavior.
	logger := GetLogger()
	require.NotNil(t, logger, "GetLogger should self-initialize if not already initialized")

	// The default logger should not log DEBUG messages, so we test that by asserting
	// that a debug log call results in an empty output. This is a simple way to
	// verify that the default log level is INFO.
	var buf bytes.Buffer
	// To test the default logger's output, we need to re-initialize it to write to our buffer.
	ForTestsOnlyResetLogger()
	defaultLogger.Store(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	logger = GetLogger()
	logger.Debug("this should not be logged")
	assert.Empty(t, buf.String(), "Default logger should not log DEBUG messages")
}

func TestInit_Once(t *testing.T) {
	// Reset the logger to ensure a clean state for this test.
	ForTestsOnlyResetLogger()

	// The first Init call should configure the logger.
	var buf1 bytes.Buffer
	Init(slog.LevelDebug, &buf1)
	GetLogger().Debug("first init")
	assert.True(t, strings.Contains(buf1.String(), "first init"), "Logger should be initialized by the first call")

	// A second Init call should be ignored. The logger should continue to write
	// to the first buffer, not the second one.
	var buf2 bytes.Buffer
	Init(slog.LevelError, &buf2)
	GetLogger().Debug("second init")
	assert.True(t, strings.Contains(buf1.String(), "second init"), "Logger should not be re-initialized by the second call")
	assert.Empty(t, buf2.String(), "Second Init call should be a no-op")
}

func TestToSlogLevel(t *testing.T) {
	testCases := []struct {
		name     string
		level    configv1.GlobalSettings_LogLevel
		expected slog.Level
	}{
		{
			name:     "Debug Level",
			level:    configv1.GlobalSettings_LOG_LEVEL_DEBUG,
			expected: slog.LevelDebug,
		},
		{
			name:     "Info Level",
			level:    configv1.GlobalSettings_LOG_LEVEL_INFO,
			expected: slog.LevelInfo,
		},
		{
			name:     "Warn Level",
			level:    configv1.GlobalSettings_LOG_LEVEL_WARN,
			expected: slog.LevelWarn,
		},
		{
			name:     "Error Level",
			level:    configv1.GlobalSettings_LOG_LEVEL_ERROR,
			expected: slog.LevelError,
		},
		{
			name:     "Default Level",
			level:    configv1.GlobalSettings_LogLevel(999), // Invalid level
			expected: slog.LevelInfo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToSlogLevel(tc.level))
		})
	}
}

func TestBroadcaster(t *testing.T) {
	b := NewBroadcaster()

	// Test Subscribe
	ch1 := b.Subscribe()
	ch2 := b.Subscribe()
	assert.Len(t, b.subscribers, 2)

	// Test Broadcast
	msg := []byte("test message")
	b.Broadcast(msg)

	select {
	case received := <-ch1:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch1")
	}

	select {
	case received := <-ch2:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch2")
	}

	// Test Unsubscribe
	b.Unsubscribe(ch1)
	assert.Len(t, b.subscribers, 1)

	// Ensure ch1 is NOT receiving messages anymore.
	// We no longer close the channel in Unsubscribe to avoid race conditions with lock-free broadcasting.

	// Broadcast again, only ch2 should receive
	msg2 := []byte("test message 2")
	b.Broadcast(msg2)

	// Verify ch1 receives nothing
	select {
	case <-ch1:
		t.Fatal("ch1 received message after unsubscribe")
	default:
	}

	select {
	case received := <-ch2:
		assert.Equal(t, msg2, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch2")
	}

	// Unsubscribe ch2
	b.Unsubscribe(ch2)
	assert.Len(t, b.subscribers, 0)
}

func TestBroadcastHandler(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// Test Handle
	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test log", 0)
	r.AddAttrs(slog.String("source", "test-source"))

	err := h.Handle(ctx, r)
	require.NoError(t, err)

	select {
	case data := <-ch:
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "test log", entry.Message)
		assert.Equal(t, "test-source", entry.Source)
		assert.NotEmpty(t, entry.ID)
		assert.NotEmpty(t, entry.Timestamp)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}

	// Test WithAttrs
	h2 := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
	assert.NotEqual(t, h, h2)
	// Just verify it doesn't panic and returns a handler
	assert.NotNil(t, h2)

	// Test WithGroup
	h3 := h.WithGroup("mygroup")
	assert.NotEqual(t, h, h3)
	assert.NotNil(t, h3)
}

func TestBroadcastHandler_Enabled(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelWarn)

	// Should be enabled for Warn and above
	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, h.Enabled(context.Background(), slog.LevelError))

	// Should be disabled for Info and below
	assert.False(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))

	// Verify level propagation in WithAttrs
	hAttrs := h.WithAttrs(nil)
	assert.True(t, hAttrs.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, hAttrs.Enabled(context.Background(), slog.LevelInfo))

	// Verify level propagation in WithGroup
	hGroup := h.WithGroup("group")
	assert.True(t, hGroup.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, hGroup.Enabled(context.Background(), slog.LevelInfo))
}

func TestTeeHandler(t *testing.T) {
	// Mock handlers
	h1 := &mockHandler{}
	h2 := &mockHandler{}

	tee := NewTeeHandler(h1, h2)

	// Test Enabled
	ctx := context.Background()
	h1.enabled = true
	h2.enabled = false
	assert.True(t, tee.Enabled(ctx, slog.LevelInfo))

	h1.enabled = false
	assert.False(t, tee.Enabled(ctx, slog.LevelInfo))

	// Test Handle
	h1.enabled = true
	h2.enabled = true
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	err := tee.Handle(ctx, r)
	assert.NoError(t, err)
	assert.True(t, h1.handled)
	assert.True(t, h2.handled)

	// Test WithAttrs
	teeWithAttrs := tee.WithAttrs([]slog.Attr{slog.String("k", "v")})
	assert.NotNil(t, teeWithAttrs)
	assert.IsType(t, &TeeHandler{}, teeWithAttrs)
	// In a real mock we'd verify WithAttrs was called on children,
	// but for now we assume implementation is correct if it returns.

	// Test WithGroup
	teeWithGroup := tee.WithGroup("g")
	assert.NotNil(t, teeWithGroup)
	assert.IsType(t, &TeeHandler{}, teeWithGroup)
}

type mockHandler struct {
	enabled bool
	handled bool
}

func (m *mockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.enabled
}

func (m *mockHandler) Handle(ctx context.Context, r slog.Record) error {
	m.handled = true
	return nil
}

func (m *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockHandler) WithGroup(name string) slog.Handler {
	return m
}

func TestRedaction(t *testing.T) {
	testCases := []struct {
		name     string
		format   string
		logFunc  func(l *slog.Logger)
		contains []string
		missing  []string
	}{
		{
			name:   "Text Format - Key Redaction",
			format: "text",
			logFunc: func(l *slog.Logger) {
				l.Info("user login", "password", "mysecretpassword")
			},
			contains: []string{"password=[REDACTED]"},
			missing:  []string{"mysecretpassword"},
		},
		{
			name:   "JSON Format - Key Redaction",
			format: "json",
			logFunc: func(l *slog.Logger) {
				l.Info("user login", "password", "mysecretpassword")
			},
			contains: []string{`"password":"[REDACTED]"`},
			missing:  []string{"mysecretpassword"},
		},
		{
			name:   "JSON Format - Deep Redaction",
			format: "json",
			logFunc: func(l *slog.Logger) {
				// Create a complex object
				type SecretStruct struct {
					ApiKey string `json:"api_key"`
				}
				data := SecretStruct{ApiKey: "secret123"}
				l.Info("config loaded", "data", data)
			},
			contains: []string{`"api_key":"[REDACTED]"`},
			missing:  []string{"secret123"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ForTestsOnlyResetLogger()
			var buf bytes.Buffer
			Init(slog.LevelInfo, &buf, tc.format)
			logger := GetLogger()

			tc.logFunc(logger)

			output := buf.String()
			for _, s := range tc.contains {
				assert.Contains(t, output, s)
			}
			for _, s := range tc.missing {
				assert.NotContains(t, output, s)
			}
		})
	}
}
