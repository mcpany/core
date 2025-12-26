// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"testing"

	"github.com/mcpany/core/pkg/upstream"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockFactory implements factory.Factory for testing
type mockFactory struct {
	NewUpstreamFunc func(config *config_v1.UpstreamServiceConfig) (upstream.Upstream, error)
}

func (m *mockFactory) NewUpstream(config *config_v1.UpstreamServiceConfig) (upstream.Upstream, error) {
	if m.NewUpstreamFunc != nil {
		return m.NewUpstreamFunc(config)
	}
	return nil, fmt.Errorf("mock factory: NewUpstreamFunc not set")
}

func TestReloadConfig_FactoryError(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()

	// We don't need UpstreamFactory mock anymore because ReloadConfig uses ServiceRegistry.
	// We mock ServiceRegistry to return an error.

	mockRegistry := new(MockServiceRegistry)
	app.serviceRegistry = mockRegistry

	configContent := `
upstream_services:
 - name: "test-service"
   http_service:
     address: "http://example.com"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	mockRegistry.On("GetAllServices").Return([]*config_v1.UpstreamServiceConfig{}, nil)
	mockRegistry.On("RegisterService", mock.Anything, mock.Anything).Return("", nil, nil, fmt.Errorf("factory error"))

	err = app.ReloadConfig(fs, []string{"/config.yaml"})
	require.NoError(t, err) // Logs error but returns nil

	mockRegistry.AssertExpectations(t)
}
