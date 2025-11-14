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
)

// setupTest creates a new cobra command, resets viper and binds flags for testing.
func setupTest(t *testing.T) (*cobra.Command, func()) {
	viper.Reset()
	cmd := &cobra.Command{
		// cobra prints errors and exits the process by default.
		// We want to handle errors in the test.
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           func(cmd *cobra.Command, args []string) {},
	}
	BindFlags(cmd)

	cleanup := func() {
		viper.Reset()
	}
	return cmd, cleanup
}

func TestBindFlags(t *testing.T) {
	cmd, cleanup := setupTest(t)
	defer cleanup()

	args := []string{
		"--jsonrpc-port=8080",
		"--config-path=/etc/mcpany",
		"--grpc-port=8081",
		"--stdio",
		"--debug",
		"--shutdown-timeout=10s",
		"--logfile=/var/log/mcpany.log",
	}
	cmd.SetArgs(args)

	assert.NoError(t, cmd.Execute())

	assert.Equal(t, "8080", viper.GetString("jsonrpc-port"))
	assert.Equal(t, []string{"/etc/mcpany"}, viper.GetStringSlice("config-path"))
	assert.Equal(t, "8081", viper.GetString("grpc-port"))
	assert.True(t, viper.GetBool("stdio"))
	assert.True(t, viper.GetBool("debug"))
	assert.Equal(t, 10*time.Second, viper.GetDuration("shutdown-timeout"))
	assert.Equal(t, "/var/log/mcpany.log", viper.GetString("logfile"))
}

func TestGetters(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	viper.Set("grpc-port", "8081")
	viper.Set("stdio", true)
	viper.Set("config-path", []string{"/etc/mcpany"})
	viper.Set("debug", true)
	viper.Set("logfile", "/var/log/mcpany.log")
	viper.Set("shutdown-timeout", 10*time.Second)

	assert.Equal(t, "8081", GetGRPCPort())
	assert.True(t, GetStdio())
	assert.Equal(t, []string{"/etc/mcpany"}, GetConfigPaths())
	assert.True(t, IsDebug())
	assert.Equal(t, "/var/log/mcpany.log", GetLogFile())
	assert.Equal(t, 10*time.Second, GetShutdownTimeout())
}

func TestGetLogLevel(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	viper.Set("debug", true)
	assert.Equal(t, v1.GlobalSettings_DEBUG, GetLogLevel())

	viper.Set("debug", false)
	assert.Equal(t, v1.GlobalSettings_INFO, GetLogLevel())
}

func TestGetBindAddress(t *testing.T) {
	t.Run("config file takes precedence when flag is not set", func(t *testing.T) {
		cmd, cleanup := setupTest(t)
		defer cleanup()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "/etc/mcpany/server.yaml", []byte(`
global_settings:
  bind_address: ":9090"
`), 0644)

		cmd.SetArgs([]string{"--config-path=/etc/mcpany"})
		assert.NoError(t, cmd.Execute())

		bindAddress, err := GetBindAddress(cmd, fs)
		assert.NoError(t, err)
		assert.Equal(t, ":9090", bindAddress)
	})

	t.Run("flag takes precedence over config file", func(t *testing.T) {
		cmd, cleanup := setupTest(t)
		defer cleanup()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "/etc/mcpany/server.yaml", []byte(`
global_settings:
  bind_address: ":9090"
`), 0644)

		cmd.SetArgs([]string{"--config-path=/etc/mcpany", "--jsonrpc-port=8080"})
		assert.NoError(t, cmd.Execute())

		bindAddress, err := GetBindAddress(cmd, fs)
		assert.NoError(t, err)
		assert.Equal(t, "8080", bindAddress)
	})

	t.Run("returns default when nothing is set", func(t *testing.T) {
		cmd, cleanup := setupTest(t)
		defer cleanup()

		fs := afero.NewMemMapFs()

		cmd.SetArgs([]string{})
		assert.NoError(t, cmd.Execute())

		bindAddress, err := GetBindAddress(cmd, fs)
		assert.NoError(t, err)
		assert.Equal(t, "50050", bindAddress) // The default value
	})
}
