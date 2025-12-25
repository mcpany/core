// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

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
