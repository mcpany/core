package config

import (
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSettings_Load(t *testing.T) {
	// Reset viper for testing
	viper.Reset()

	// Setup fs
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	// Set viper values
	viper.Set("grpc-port", "50051")
	viper.Set("stdio", true)
	viper.Set("config-path", []string{"/config.yaml"})
	viper.Set("debug", true)
	viper.Set("log-level", "debug")
	// Use temp file for log
	tmpLog, err := os.CreateTemp("", "app.log")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpLog.Name()) }()
	_ = tmpLog.Close()

	viper.Set("logfile", tmpLog.Name())
	viper.Set("shutdown-timeout", 5*time.Second)
	viper.Set("profiles", []string{"profile1", "profile2"})
	viper.Set("mcp-listen-address", "127.0.0.1:8080")
	viper.Set("api-key", "global-key")
	viper.Set("metrics-listen-address", "127.0.0.1:9090")

	// Create a dummy config file
	err = afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services: []
`), 0o644)
	require.NoError(t, err)

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err = settings.Load(cmd, fs)
	require.NoError(t, err)

	// Verify values
	assert.Equal(t, "50051", settings.GRPCPort())
	assert.True(t, settings.Stdio())
	assert.Equal(t, []string{"/config.yaml"}, settings.ConfigPaths())
	assert.True(t, settings.IsDebug())
	assert.Equal(t, tmpLog.Name(), settings.LogFile())
	assert.Equal(t, 5*time.Second, settings.ShutdownTimeout())
	assert.Equal(t, []string{"profile1", "profile2"}, settings.Profiles())
	assert.Equal(t, "127.0.0.1:8080", settings.MCPListenAddress())
	assert.Equal(t, "global-key", settings.APIKey())
	assert.Equal(t, "127.0.0.1:9090", settings.MetricsListenAddress())
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, settings.LogLevel())
}

func TestSettings_Defaults(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err := settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, []string{"default"}, settings.Profiles())
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, settings.LogLevel())
}

func TestSettings_SetAPIKey(t *testing.T) {
	// Test SetAPIKey on a fresh instance
	// Reset singleton if possible, but we can just use new instance for safety in unit test
	// But `GlobalSettings` returns singleton.
	// Let's use a fresh instance.
	s := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	s.SetAPIKey("new-key")
	assert.Equal(t, "new-key", s.APIKey())
}

func TestSettings_LoggingInit(t *testing.T) {
	// Test creating a real log file
	viper.Reset()
	fs := afero.NewMemMapFs() // Used for config, but log file uses os.Open...
	// Settings.Load uses os.OpenFile, so we need real FS or patch it.
	// Since we are running in a container, we can use a temp file.
	tmpFile, err := os.CreateTemp("", "test-log")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close() // Close it so Load can open it

	viper.Set("logfile", tmpFile.Name())

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	// We can't Mock os.OpenFile easily without separate function.
	// So we tested with a real file path.
	err = settings.Load(&cobra.Command{}, fs)
	require.NoError(t, err)
	assert.Equal(t, tmpFile.Name(), settings.LogFile())
}

func TestSettings_MCPListenAddress_Precedence(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	viper.Set("config-path", []string{"/config.yaml"})
	viper.Set("mcp-listen-address", "flag-value") // Should be overridden if config present?
	// The code says: if NOT Changed("mcp-listen-address") AND config has it -> use config.
	// But viper.Set sets it, doesn't mark as changed in cmd.Flags().
	// cmd.Flags().Changed check is on Cobra flags.

	// Write config with mcp address
	configContent := `
global_settings:
  mcp_listen_address: "127.0.0.1:9091"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err = settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1:9091", settings.MCPListenAddress())
}

func TestSettings_MCPListenAddress_EnvPrecedence(t *testing.T) {
	// This test verifies that Environment Variables override Config Files.
	// Precedence should be: Flag > Env > Config > Default
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	// 1. Set Environment Variable
	envKey := "MCPANY_MCP_LISTEN_ADDRESS"
	expectedVal := "127.0.0.1:1111"
	t.Setenv(envKey, expectedVal)

	// 2. Set Config File
	viper.Set("config-path", []string{"/config.yaml"})
	configContent := `
global_settings:
  mcp_listen_address: "127.0.0.1:9999"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	// We need to ensure viper reads the env var.
	// Since we can't easily re-bind flags in this test context without full setup,
	// we rely on viper.AutomaticEnv() being called in BindRootFlags, but here we invoke Load directly.
	// We need to mimic what BindRootFlags does for viper.
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MCPANY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Note: settings.Load uses viper.GetString("mcp-listen-address").
	// viper.GetString will pick up the Env var if set.

	err = settings.Load(cmd, fs)
	require.NoError(t, err)

	// Expect Env Value (1111), NOT Config Value (9999)
	assert.Equal(t, expectedVal, settings.MCPListenAddress())
}

func TestSettings_GetDbDsn(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	viper.Set("db-dsn", "postgres://user:pass@127.0.0.1:5432/db")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err := settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, "postgres://user:pass@127.0.0.1:5432/db", settings.GetDbDsn())
}

func TestSettings_GetDbDriver(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	viper.Set("db-driver", "postgres")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err := settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, "postgres", settings.GetDbDriver())
}

func TestSettings_GetDlp(t *testing.T) {
	enabled := true
	settings := &Settings{
		proto: func() *configv1.GlobalSettings {
			dlp := configv1.DLPConfig_builder{
				Enabled: proto.Bool(enabled),
			}.Build()
			return configv1.GlobalSettings_builder{
				Dlp: dlp,
			}.Build()
		}(),
	}
	assert.NotNil(t, settings.GetDlp())
	assert.True(t, settings.GetDlp().GetEnabled())
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestSettings_ExtraGetters(t *testing.T) {
	// Create a Settings instance manually with populated fields
	middlewares := []*configv1.Middleware{
		func() *configv1.Middleware {
			return configv1.Middleware_builder{
				Name: proto.String("test-middleware"),
			}.Build()
		}(),
	}

	s := &Settings{
		dbPath: "/path/to/db.sqlite",
		proto: func() *configv1.GlobalSettings {
			return configv1.GlobalSettings_builder{
				Middlewares: middlewares,
			}.Build()
		}(),
	}

	assert.Equal(t, "/path/to/db.sqlite", s.DBPath())
	assert.Equal(t, middlewares, s.Middlewares())
}
