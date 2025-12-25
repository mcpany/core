// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
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
	mockFactory := &mockFactory{
		NewUpstreamFunc: func(_ *config_v1.UpstreamServiceConfig) (upstream.Upstream, error) {
			return nil, fmt.Errorf("factory error")
		},
	}
	app.UpstreamFactory = mockFactory

	// Initialize ServiceRegistry as it is now required by ReloadConfig
	app.ServiceRegistry = serviceregistry.New(
		mockFactory,
		tool.NewManager(nil),
		prompt.NewManager(),
		resource.NewManager(),
		auth.NewManager(),
	)

	configContent := `
upstream_services:
 - name: "test-service"
   http_service:
     address: "http://example.com"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	err = app.ReloadConfig(fs, []string{"/config.yaml"})
	require.NoError(t, err) // Logs error but returns nil
}
