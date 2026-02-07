// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type hunterMockUpstream struct {
	upstream.Upstream
}

func (m *hunterMockUpstream) Register(_ context.Context, _ *configv1.UpstreamServiceConfig, _ tool.ManagerInterface, _ prompt.ManagerInterface, _ resource.ManagerInterface, _ bool) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "mock-service-key", nil, nil, nil
}
func (m *hunterMockUpstream) Shutdown(_ context.Context) error { return nil }

type hunterMockFactory struct {
	factory.Factory
}

func (m *hunterMockFactory) NewUpstream(_ *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	return &hunterMockUpstream{}, nil
}

type hunterMockToolManager struct {
	tool.ManagerInterface
}
func (m *hunterMockToolManager) AddTool(_ tool.Tool) error             { return nil }
func (m *hunterMockToolManager) ClearToolsForService(_ string)         {}
func (m *hunterMockToolManager) GetTool(_ string) (tool.Tool, bool)    { return nil, false }
func (m *hunterMockToolManager) ListTools() []tool.Tool                { return nil }
func (m *hunterMockToolManager) ListServices() []*tool.ServiceInfo     { return nil }
func (m *hunterMockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}
func (m *hunterMockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return nil, nil }
func (m *hunterMockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func TestServiceRegistry_SecretsLeak(t *testing.T) {
	f := &hunterMockFactory{}
	tm := &hunterMockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("secret-service")
	serviceConfig.SetAuthentication(configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			VerificationValue: proto.String("SUPER_SECRET_VALUE"),
		}.Build(),
	}.Build())

	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Get the config via GetServiceConfig
	retrievedConfig, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)

	auth := retrievedConfig.GetAuthentication()
	require.NotNil(t, auth)
	apiKey := auth.GetApiKey()
	require.NotNil(t, apiKey)

	// This should fail currently
	assert.Empty(t, apiKey.GetVerificationValue(), "API Key secret should be scrubbed in GetServiceConfig")

	// Get all services
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	require.Len(t, services, 1)

	authAll := services[0].GetAuthentication()
	require.NotNil(t, authAll)
	apiKeyAll := authAll.GetApiKey()
	// This should fail currently
	assert.Empty(t, apiKeyAll.GetVerificationValue(), "API Key secret should be scrubbed in GetAllServices")
}
