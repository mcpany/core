package config

import (
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	viper.Set("mcp-listen-address", "localhost:8080")
	viper.Set("api-key", "global-key")
	viper.Set("metrics-listen-address", "localhost:9090")

	// Create a dummy config file
	err = afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services: []
`), 0o644)
	require.NoError(t, err)

	settings := &Settings{
		proto: &configv1.GlobalSettings{},
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
	assert.Equal(t, "localhost:8080", settings.MCPListenAddress())
	assert.Equal(t, "global-key", settings.APIKey())
	assert.Equal(t, "localhost:9090", settings.MetricsListenAddress())
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, settings.LogLevel())
}

func TestSettings_Defaults(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	settings := &Settings{
		proto: &configv1.GlobalSettings{},
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
		proto: &configv1.GlobalSettings{},
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
		proto: &configv1.GlobalSettings{},
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
  mcp_listen_address: "localhost:9091"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	settings := &Settings{
		proto: &configv1.GlobalSettings{},
	}
	err = settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, "localhost:9091", settings.MCPListenAddress())
}
