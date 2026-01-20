package app

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRun_AutoEnableFileConfig(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create a malformed config file.
	// If file config is enabled, loading this will cause an error.
	// If file config is IGNORED (old behavior), this would be skipped and startup would succeed (using mock DB).
	err := afero.WriteFile(fs, "/config.yaml", []byte("malformed: :"), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	app := NewApplication()
	mockStore := new(MockStore)
	mockStore.On("Load", mock.Anything).Return((*configv1.McpAnyServerConfig)(nil), nil)
	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{}, nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return(&configv1.GlobalSettings{}, nil)
	mockStore.On("Close").Return(nil)
	app.Storage = mockStore

	// Ensure env var is unset/false
	t.Setenv("MCPANY_ENABLE_FILE_CONFIG", "false")

	// We run with a config path.
	err = app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, "", 5*time.Second)

	// It SHOULD error because it tried to load the malformed file.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "yaml", "Should fail due to yaml parsing error, indicating file config was enabled")
}
