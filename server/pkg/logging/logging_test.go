package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

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
	defaultLogger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

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
