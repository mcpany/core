package config

import (
	"io"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettings_StdioLogging(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	viper.Set("stdio", true)
	// Ensure no logfile is set, so it falls back to stdio logic
	viper.Set("logfile", "")

	// Reset logger to ensure Init is called effectively
	logging.ForTestsOnlyResetLogger()

	// Capture stderr
	r, w, _ := os.Pipe()
	originalStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = originalStderr }()

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err := settings.Load(cmd, fs)
	require.NoError(t, err)

	// Log something using the global logger
	logging.GetLogger().Info("test log entry")

	// Close writer to flush and read
	_ = w.Close()
	out, _ := io.ReadAll(r)

	// Assert that the log entry was written to stderr
	// In the old code (io.Discard), this would fail.
	assert.Contains(t, string(out), "test log entry")
}
