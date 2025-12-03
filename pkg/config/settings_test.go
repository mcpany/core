/*
 * Copyright 2024 Author(s) of MCP Any
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

package config

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mcpany/core/pkg/logging"
	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogLevel(t *testing.T) {
	testCases := []struct {
		name          string
		logLevel      string
		debug         bool
		expectedLevel v1.GlobalSettings_LogLevel
		expectedLog   string
	}{
		{
			name:          "Empty log level",
			logLevel:      "",
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_INFO,
			expectedLog:   "Invalid log level specified: ''. Defaulting to INFO.",
		},
		{
			name:          "Invalid log level",
			logLevel:      "invalid",
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_INFO,
			expectedLog:   "Invalid log level specified: 'invalid'. Defaulting to INFO.",
		},
		{
			name:          "Debug log level",
			logLevel:      "debug",
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_DEBUG,
			expectedLog:   "",
		},
		{
			name:          "Info log level",
			logLevel:      "info",
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_INFO,
			expectedLog:   "",
		},
		{
			name:          "Warn log level",
			logLevel:      "warn",
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_WARN,
			expectedLog:   "",
		},
		{
			name:          "Error log level",
			logLevel:      "error",
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_ERROR,
			expectedLog:   "",
		},
		{
			name:          "Debug flag overrides log level",
			logLevel:      "info",
			debug:         true,
			expectedLevel: v1.GlobalSettings_LOG_LEVEL_DEBUG,
			expectedLog:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logging.ForTestsOnlyResetLogger()
			var logBuffer bytes.Buffer
			logging.Init(slog.LevelDebug, &logBuffer)

			viper.Set("log-level", tc.logLevel)
			viper.Set("debug", tc.debug)

			settings := GlobalSettings()
			settings.logLevel = tc.logLevel
			settings.debug = tc.debug
			logLevel := settings.LogLevel()

			assert.Equal(t, tc.expectedLevel, logLevel)
			if tc.expectedLog != "" {
				assert.Contains(t, logBuffer.String(), tc.expectedLog)
			}
		})
	}
}
