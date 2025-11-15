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

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	v1 "github.com/mcpany/core/proto/config/v1"
)

func TestGlobalSettings(t *testing.T) {
	settings1 := GlobalSettings()
	settings2 := GlobalSettings()
	assert.Same(t, settings1, settings2, "GlobalSettings should return a singleton instance")
}

func TestSettings_LoadAndGetters(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	configDir := "/etc/mcpany"
	configFile := configDir + "/server.yaml"
	assert.NoError(t, fs.MkdirAll(configDir, 0755))

	configContent := []byte(`
global_settings:
  mcp_listen_address: "from-file:5000"
`)
	assert.NoError(t, afero.WriteFile(fs, configFile, configContent, 0644))

	cmd := &cobra.Command{}
	cmd.Flags().String("grpc-port", "50051", "")
	cmd.Flags().Bool("stdio", false, "")
	cmd.Flags().StringSlice("config-path", []string{}, "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().String("log-level", "info", "")
	cmd.Flags().String("logfile", "", "")
	cmd.Flags().Duration("shutdown-timeout", 10*time.Second, "")
	cmd.Flags().String("mcp-listen-address", "localhost:1234", "")

	viper.AutomaticEnv()
	assert.NoError(t, viper.BindPFlags(cmd.Flags()))

	viper.Set("grpc-port", "8080")
	viper.Set("stdio", true)
	viper.Set("config-path", []string{configDir})
	viper.Set("debug", true)
	viper.Set("log-level", "debug")
	viper.Set("logfile", "/var/log/mcpany.log")
	viper.Set("shutdown-timeout", 20*time.Second)

	settings := GlobalSettings()
	err := settings.Load(cmd, fs)
	assert.NoError(t, err)

	assert.Equal(t, "8080", settings.GRPCPort())
	assert.True(t, settings.Stdio())
	assert.Equal(t, []string{configDir}, settings.ConfigPaths())
	assert.True(t, settings.IsDebug())
	assert.Equal(t, "/var/log/mcpany.log", settings.LogFile())
	assert.Equal(t, 20*time.Second, settings.ShutdownTimeout())
	assert.Equal(t, "from-file:5000", settings.MCPListenAddress())
	assert.Equal(t, v1.GlobalSettings_DEBUG, settings.LogLevel())
}

func TestSettings_LogLevel(t *testing.T) {
	tests := []struct {
		name     string
		debug    bool
		logLevel string
		expected v1.GlobalSettings_LogLevel
	}{
		{"debug enabled", true, "info", v1.GlobalSettings_DEBUG},
		{"debug level", false, "debug", v1.GlobalSettings_DEBUG},
		{"info level", false, "info", v1.GlobalSettings_INFO},
		{"warn level", false, "warn", v1.GlobalSettings_WARN},
		{"error level", false, "error", v1.GlobalSettings_ERROR},
		{"invalid level", false, "invalid", v1.GlobalSettings_INFO},
		{"uppercase level", false, "ERROR", v1.GlobalSettings_ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Settings{
				debug:    tt.debug,
				logLevel: tt.logLevel,
			}
			assert.Equal(t, tt.expected, s.LogLevel())
		})
	}
}
