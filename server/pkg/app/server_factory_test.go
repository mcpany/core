// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"testing"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/upstream"
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
	app.UpstreamFactory = &mockFactory{
		NewUpstreamFunc: func(_ *config_v1.UpstreamServiceConfig) (upstream.Upstream, error) {
			return nil, fmt.Errorf("factory error")
		},
	}

	configContent := `
upstream_services:
 - name: "test-service"
   http_service:
     address: "http://example.com"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	err = app.ReloadConfig(context.Background(), fs, []string{"/config.yaml"})
	require.NoError(t, err) // Logs error but returns nil
}
