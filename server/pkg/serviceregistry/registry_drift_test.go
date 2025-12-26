package serviceregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistry_DriftFix(t *testing.T) {
	upstreamErr := errors.New("registration failed")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(_ string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					return "", nil, nil, upstreamErr
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("broken-service")

	// 1. Register should fail
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), upstreamErr.Error())

	// 2. But service should be listed (Fix Config Drift)
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, "broken-service", services[0].GetName())

	// 3. And error should be available
	serviceID, err := util.SanitizeServiceName("broken-service")
	require.NoError(t, err)

	errorMsg, ok := registry.GetServiceError(serviceID)
	require.True(t, ok, "Error should be stored")
	assert.Contains(t, errorMsg, upstreamErr.Error())

	// 4. Config should be retrievable
	cfg, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, serviceConfig, cfg)
}
