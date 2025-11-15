/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"bytes"
	"log/slog"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestToSlogLevel(t *testing.T) {
	testCases := []struct {
		name     string
		level    configv1.GlobalSettings_LogLevel
		expected slog.Level
	}{
		{
			name:     "Debug",
			level:    configv1.GlobalSettings_LOG_LEVEL_DEBUG,
			expected: slog.LevelDebug,
		},
		{
			name:     "Info",
			level:    configv1.GlobalSettings_LOG_LEVEL_INFO,
			expected: slog.LevelInfo,
		},
		{
			name:     "Warn",
			level:    configv1.GlobalSettings_LOG_LEVEL_WARN,
			expected: slog.LevelWarn,
		},
		{
			name:     "Error",
			level:    configv1.GlobalSettings_LOG_LEVEL_ERROR,
			expected: slog.LevelError,
		},
		{
			name:     "Default",
			level:    configv1.GlobalSettings_LogLevel(999),
			expected: slog.LevelInfo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToSlogLevel(tc.level))
		})
	}
}

func TestGetLogger(t *testing.T) {
	ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	defaultLogger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	once = sync.Once{}
	once.Do(func() {})

	logger := GetLogger()
	assert.NotNil(t, logger)
	logger.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestInit(t *testing.T) {
	ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	Init(slog.LevelDebug, &buf)

	logger := GetLogger()
	logger.Debug("test debug message")
	assert.Contains(t, buf.String(), "test debug message")
}
