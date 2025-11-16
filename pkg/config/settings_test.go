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

package config

import (
	"testing"
	"time"

	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalSettings(t *testing.T) {
	s1 := GlobalSettings()
	s2 := GlobalSettings()
	assert.Same(t, s1, s2, "GlobalSettings should return a singleton")
}

func TestLoad(t *testing.T) {
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}
	cmd.Flags().String("grpc-port", "50051", "")
	cmd.Flags().Bool("stdio", false, "")
	cmd.Flags().StringSlice("config-path", []string{}, "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().String("log-level", "info", "")
	cmd.Flags().String("logfile", "", "")
	cmd.Flags().Duration("shutdown-timeout", 10*time.Second, "")
	cmd.Flags().String("mcp-listen-address", "localhost:50050", "")

	viper.BindPFlags(cmd.Flags())

	s := GlobalSettings()
	err := s.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, "50051", s.GRPCPort())
	assert.False(t, s.Stdio())
	assert.Equal(t, []string{}, s.ConfigPaths())
	assert.False(t, s.IsDebug())
	assert.Equal(t, "", s.LogFile())
	assert.Equal(t, 10*time.Second, s.ShutdownTimeout())
	assert.Equal(t, "localhost:50050", s.MCPListenAddress())
	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_INFO, s.LogLevel())
}

func TestLogLevels(t *testing.T) {
	s := &Settings{}

	s.logLevel = "debug"
	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_DEBUG, s.LogLevel())

	s.logLevel = "info"
	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_INFO, s.LogLevel())

	s.logLevel = "warn"
	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_WARN, s.LogLevel())

	s.logLevel = "error"
	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_ERROR, s.LogLevel())

	s.logLevel = "invalid"
	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_INFO, s.LogLevel())
}
