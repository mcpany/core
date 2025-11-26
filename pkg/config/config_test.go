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
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/mcpany/core/pkg/logging"
	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogLevel_InvalidLevelWarning(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	logging.Init(slog.LevelInfo, &buf)

	v := viper.New()
	v.Set("log-level", "invalid-level")
	v.Set("debug", false)

	settings := &Settings{v: v}
	logLevel := settings.LogLevel()

	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_INFO, logLevel)
	logs := buf.String()
	assert.Contains(t, logs, "Invalid log level specified: 'invalid-level'. Defaulting to INFO.")
}

func TestBindFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindFlags(cmd)

	// Check a fiew flags to ensure they are bound
	assert.NotNil(t, cmd.PersistentFlags().Lookup("mcp-listen-address"))
	assert.NotNil(t, cmd.Flags().Lookup("grpc-port"))
	assert.NotNil(t, cmd.Flags().Lookup("stdio"))

	// Check that the values are correctly bound to viper
	cmd.Flags().Set("grpc-port", "8081")
	assert.Equal(t, "8081", viper.GetString("grpc-port"))

	cmd.Flags().Set("stdio", "true")
	assert.True(t, viper.GetBool("stdio"))
}

func TestGRPCPortEnvVar(t *testing.T) {
	viper.Reset() // Reset viper to avoid state leakage from other tests.
	os.Setenv("MCPANY_GRPC_PORT", "9090")
	defer os.Unsetenv("MCPANY_GRPC_PORT")

	cmd := &cobra.Command{}
	BindFlags(cmd)

	assert.Equal(t, "9090", viper.GetString("grpc-port"))
}

func TestMCPListenAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "port only",
			address:  "50050",
			expected: "localhost:50050",
		},
		{
			name:     "address with port",
			address:  "127.0.0.1:50050",
			expected: "127.0.0.1:50050",
		},
		{
			name:     "hostname with port",
			address:  "mcpany.internal:50050",
			expected: "mcpany.internal:50050",
		},
		{
			name:     "hostname without port",
			address:  "mcpany.internal",
			expected: "localhost:mcpany.internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.Set("mcp-listen-address", tt.address)
			settings := &Settings{v: v}
			assert.Equal(t, tt.expected, settings.MCPListenAddress())
		})
	}
}
