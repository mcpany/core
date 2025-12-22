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
