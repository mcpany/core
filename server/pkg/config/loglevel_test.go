package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestLogLevel_CaseSensitivity(t *testing.T) {
	// Note: We use the global settings instance here as the Settings struct
    // is designed as a singleton. In a real scenario, we should ensure tests
    // run sequentially or mock the settings if possible to avoid side effects.
    // However, for this specific test, we are modifying and checking the
    // logLevel field which is sufficient for verifying the parsing logic.
	s := GlobalSettings()

	tests := []struct {
		name     string
		input    string
		expected configv1.GlobalSettings_LogLevel
	}{
		{
			name:     "Lower case info",
			input:    "info",
			expected: configv1.GlobalSettings_LOG_LEVEL_INFO,
		},
		{
			name:     "Upper case INFO",
			input:    "INFO",
			expected: configv1.GlobalSettings_LOG_LEVEL_INFO,
		},
		{
			name:     "Lower case warning",
			input:    "warning",
			expected: configv1.GlobalSettings_LOG_LEVEL_WARN,
		},
		{
			name:     "Upper case WARNING",
			input:    "WARNING",
			expected: configv1.GlobalSettings_LOG_LEVEL_WARN,
		},
		{
			name:     "Lower case error",
			input:    "error",
			expected: configv1.GlobalSettings_LOG_LEVEL_ERROR,
		},
        {
			name:     "Lower case warn",
			input:    "warn",
			expected: configv1.GlobalSettings_LOG_LEVEL_WARN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.logLevel = tt.input
			s.debug = false
			assert.Equal(t, tt.expected, s.LogLevel(), "LogLevel() should return expected enum for input %s", tt.input)
		})
	}
}
