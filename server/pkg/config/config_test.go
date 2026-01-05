// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/logging"
	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogLevel_InvalidLevelWarning(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	logging.Init(slog.LevelInfo, &buf)

	settings := &Settings{
		logLevel: "invalid-level",
		debug:    false,
	}

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
	_ = cmd.Flags().Set("grpc-port", "8081")
	assert.Equal(t, "8081", viper.GetString("grpc-port"))

	_ = cmd.Flags().Set("stdio", "true")
	assert.True(t, viper.GetBool("stdio"))
}

func TestGRPCPortEnvVar(t *testing.T) {
	viper.Reset() // Reset viper to avoid state leakage from other tests.
	_ = os.Setenv("MCPANY_GRPC_PORT", "9090")
	defer func() { _ = os.Unsetenv("MCPANY_GRPC_PORT") }()

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
			expected: "50050",
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
			expected: "mcpany.internal",
		},
		{
			name:     "empty address",
			address:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			cmd := &cobra.Command{}
			BindFlags(cmd)
			viper.Set("mcp-listen-address", tt.address)
			s := GlobalSettings()
			err := s.Load(cmd, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, s.MCPListenAddress())
		})
	}
}

func TestGlobalSettings(t *testing.T) {
	// To prevent test pollution, we reset viper and clear any environment variables
	// that might affect the test.
	viper.Reset()
	os.Clearenv()
	defer os.Clearenv()

	cmd := &cobra.Command{}
	BindFlags(cmd)

	// Test with default values first
	s := GlobalSettings()
	err := s.Load(cmd, nil)
	assert.NoError(t, err)

	assert.Equal(t, "", s.GRPCPort())
	assert.Equal(t, "50050", s.MCPListenAddress())
	assert.False(t, s.IsDebug())
	assert.False(t, s.Stdio())

	// Test with values from viper
	viper.Set("grpc-port", "6001")
	viper.Set("mcp-listen-address", "0.0.0.0:6000")
	viper.Set("debug", true)
	viper.Set("stdio", true)

	// Reload settings to apply viper changes
	err = s.Load(cmd, nil)
	assert.NoError(t, err)

	assert.Equal(t, "6001", s.GRPCPort())
	assert.Equal(t, "0.0.0.0:6000", s.MCPListenAddress())
	assert.True(t, s.IsDebug())
	assert.True(t, s.Stdio())
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestBindRootFlags_Comprehensive(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{}
	BindRootFlags(cmd)

	tests := []struct {
		flag     string
		defVal   interface{}
		setVal   string
		envVar   string
		envVal   string
		checkVal interface{}
	}{
		{
			flag:     "mcp-listen-address",
			defVal:   "50050",
			setVal:   "9000",
			envVar:   "MCPANY_MCP_LISTEN_ADDRESS",
			envVal:   "9999",
			checkVal: "9999",
		},
		{
			flag:     "metrics-listen-address",
			defVal:   "",
			setVal:   ":8080",
			envVar:   "MCPANY_METRICS_LISTEN_ADDRESS",
			envVal:   ":8081",
			checkVal: ":8081",
		},
		{
			flag:     "debug",
			defVal:   false,
			setVal:   "true",
			envVar:   "MCPANY_DEBUG",
			envVal:   "true",
			checkVal: "true", // viper returns string "true" for env vars if not explicitly cast
		},
		{
			flag:     "log-level",
			defVal:   "info",
			setVal:   "error",
			envVar:   "MCPANY_LOG_LEVEL",
			envVal:   "warn",
			checkVal: "warn",
		},
		{
			flag:     "log-format",
			defVal:   "text",
			setVal:   "json",
			envVar:   "MCPANY_LOG_FORMAT",
			envVal:   "json",
			checkVal: "json",
		},
		{
			flag:     "logfile",
			defVal:   "",
			setVal:   "/tmp/log",
			envVar:   "",
			envVal:   "",
			checkVal: "/tmp/log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			// Check default value
			f := cmd.PersistentFlags().Lookup(tt.flag)
			assert.NotNil(t, f)
			// viper.Get might return default from flag binding

			// Check Env override
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				// We need to re-bind or rely on viper.AutomaticEnv() which is called in BindRootFlags
				// But viper needs to see the env var when we call Get
				// Also SetEnvKeyReplacer
				// Since BindRootFlags calls AutomaticEnv and SetEnvKeyReplacer, it should work IF the env var is set before?
				// Actually viper checks env vars at the time of Get call.

				defer os.Unsetenv(tt.envVar)
				assert.Equal(t, tt.checkVal, viper.GetString(tt.flag))
			} else {
				// Check Set value
				viper.Set(tt.flag, tt.setVal)
				assert.Equal(t, tt.checkVal, viper.GetString(tt.flag))
			}
		})
	}
}

func TestBindServerFlags_Comprehensive(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{}
	// We must call BindRootFlags because it sets up AutomaticEnv and KeyReplacer!
	BindRootFlags(cmd)
	BindServerFlags(cmd)

	tests := []struct {
		flag     string
		envVar   string
		envVal   string
		checkVal interface{}
	}{
		{
			flag:     "grpc-port",
			envVar:   "MCPANY_GRPC_PORT",
			envVal:   "50051",
			checkVal: "50051",
		},
		{
			flag:     "stdio",
			envVar:   "MCPANY_STDIO",
			envVal:   "true",
			checkVal: "true",
		},
		{
			flag:     "shutdown-timeout",
			envVar:   "MCPANY_SHUTDOWN_TIMEOUT",
			envVal:   "10s",
			checkVal: "10s",
		},
		{
			flag:     "api-key",
			envVar:   "MCPANY_API_KEY",
			envVal:   "secret",
			checkVal: "secret",
		},
		{
			flag:     "db-path",
			envVar:   "MCPANY_DB_PATH",
			envVal:   "test.db",
			checkVal: "test.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
				assert.Equal(t, tt.checkVal, viper.GetString(tt.flag))
			}
		})
	}

	// Test Profiles Slice
	t.Run("profiles", func(t *testing.T) {
		os.Setenv("MCPANY_PROFILES", "p1,p2")
		defer os.Unsetenv("MCPANY_PROFILES")
		// StringSlice from env var might be space separated or handled differently by viper depending on version/config
		// Let's check string first
		assert.Equal(t, "p1,p2", viper.GetString("profiles"))
	})
}
