// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks Duplicated for Unhealthy Service Test ---
// Ideally these would be in a shared testutil package, but for now we duplicate.

type mockResourceForRepro struct {
	res     *mcp.Resource
	service string
}

func (m *mockResourceForRepro) Resource() *mcp.Resource {
	return m.res
}

func (m *mockResourceForRepro) Service() string {
	return m.service
}

func (m *mockResourceForRepro) Read(_ context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: m.res.URI, Text: "content"},
		},
	}, nil
}

func (m *mockResourceForRepro) Subscribe(_ context.Context) error {
	return nil
}

type mockResourceManagerForRepro struct {
	resources []resource.Resource
}

func (m *mockResourceManagerForRepro) ListResources() []resource.Resource {
	return m.resources
}

func (m *mockResourceManagerForRepro) GetResource(uri string) (resource.Resource, bool) {
	for _, r := range m.resources {
		if r.Resource().URI == uri {
			return r, true
		}
	}
	return nil, false
}

func (m *mockResourceManagerForRepro) AddResource(r resource.Resource) {
	m.resources = append(m.resources, r)
}

func (m *mockResourceManagerForRepro) RemoveResource(_ string) {}
func (m *mockResourceManagerForRepro) OnListChanged(f func())    {}
func (m *mockResourceManagerForRepro) ClearResourcesForService(_ string) {}

type mockPromptForRepro struct {
	p       *mcp.Prompt
	service string
}

func (m *mockPromptForRepro) Prompt() *mcp.Prompt {
	return m.p
}

func (m *mockPromptForRepro) Service() string {
	return m.service
}

func (m *mockPromptForRepro) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{},
	}, nil
}

type mockPromptManagerForRepro struct {
	prompts []prompt.Prompt
}

func (m *mockPromptManagerForRepro) ListPrompts() []prompt.Prompt {
	return m.prompts
}

func (m *mockPromptManagerForRepro) GetPrompt(name string) (prompt.Prompt, bool) {
	for _, p := range m.prompts {
		if p.Prompt().Name == name {
			return p, true
		}
	}
	return nil, false
}

func (m *mockPromptManagerForRepro) AddPrompt(p prompt.Prompt) {
	m.prompts = append(m.prompts, p)
}

func (m *mockPromptManagerForRepro) UpdatePrompt(p prompt.Prompt) {
	m.prompts = append(m.prompts, p)
}

func (m *mockPromptManagerForRepro) RemovePrompt(_ string) {}
func (m *mockPromptManagerForRepro) SetMCPServer(_ prompt.MCPServerProvider) {}
func (m *mockPromptManagerForRepro) ClearPromptsForService(_ string) {}

type serviceInfoProviderToolManagerForRepro struct {
	tool.Manager
	services map[string]*tool.ServiceInfo
}

func (m *serviceInfoProviderToolManagerForRepro) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	s, ok := m.services[id]
	return s, ok
}

// Stubs
func (m *serviceInfoProviderToolManagerForRepro) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *serviceInfoProviderToolManagerForRepro) GetTool(_ string) (tool.Tool, bool)           { return nil, false }
func (m *serviceInfoProviderToolManagerForRepro) ListTools() []tool.Tool                       { return nil }
func (m *serviceInfoProviderToolManagerForRepro) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *serviceInfoProviderToolManagerForRepro) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *serviceInfoProviderToolManagerForRepro) SetMCPServer(_ tool.MCPServerProvider)    {}
func (m *serviceInfoProviderToolManagerForRepro) AddTool(_ tool.Tool) error                { return nil }
func (m *serviceInfoProviderToolManagerForRepro) ClearToolsForService(_ string)            {}
func (m *serviceInfoProviderToolManagerForRepro) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *serviceInfoProviderToolManagerForRepro) IsServiceAllowed(_, _ string) bool                       { return true }
func (m *serviceInfoProviderToolManagerForRepro) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	return nil, false
}

func TestUnhealthyServiceResourceAccess(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Services config
	srvHealthy := &tool.ServiceInfo{
		Config:       &configv1.UpstreamServiceConfig{},
		HealthStatus: consts.HealthStatusHealthy,
	}
	srvUnhealthy := &tool.ServiceInfo{
		Config:       &configv1.UpstreamServiceConfig{},
		HealthStatus: consts.HealthStatusUnhealthy,
	}

	tm := &serviceInfoProviderToolManagerForRepro{
		services: map[string]*tool.ServiceInfo{
			"healthy-service":   srvHealthy,
			"unhealthy-service": srvUnhealthy,
		},
	}

	rm := &mockResourceManagerForRepro{
		resources: []resource.Resource{
			&mockResourceForRepro{res: &mcp.Resource{URI: "healthy://res"}, service: "healthy-service"},
			&mockResourceForRepro{res: &mcp.Resource{URI: "unhealthy://res"}, service: "unhealthy-service"},
			&mockResourceForRepro{res: &mcp.Resource{URI: "unknown://res"}, service: "unknown-service"},
		},
	}
	pm := &mockPromptManagerForRepro{}
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// 1. Verify ListResources via Middleware
	res, err := server.ResourceListFilteringMiddleware(next)(ctx, consts.MethodResourcesList, &mcp.ListResourcesRequest{})
	require.NoError(t, err)
	lRes, ok := res.(*mcp.ListResourcesResult)
	require.True(t, ok)

	foundURIs := make(map[string]bool)
	for _, r := range lRes.Resources {
		foundURIs[r.URI] = true
	}

	// Expectation: Unhealthy resource should NOT be listed
	assert.Contains(t, foundURIs, "healthy://res")
	assert.NotContains(t, foundURIs, "unhealthy://res")

	// 2. Verify ReadResource
	// Try reading unhealthy resource
	_, err = server.ReadResource(ctx, &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{
		URI: "unhealthy://res",
	}})
	// Expectation: Should fail
	require.Error(t, err)
	assert.Contains(t, err.Error(), consts.HealthStatusUnhealthy)

	// Try reading healthy resource
	resRead, err := server.ReadResource(ctx, &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{
		URI: "healthy://res",
	}})
	require.NoError(t, err)
	require.Len(t, resRead.Contents, 1)
	assert.Equal(t, "content", resRead.Contents[0].Text)

	// Try reading resource from unknown service (should succeed as default allow or skip check)
	resUnknown, err := server.ReadResource(ctx, &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{
		URI: "unknown://res",
	}})
	require.NoError(t, err)
	require.Len(t, resUnknown.Contents, 1)
}

func TestUnhealthyServicePromptAccess(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	srvHealthy := &tool.ServiceInfo{
		Config:       &configv1.UpstreamServiceConfig{},
		HealthStatus: consts.HealthStatusHealthy,
	}
	srvUnhealthy := &tool.ServiceInfo{
		Config:       &configv1.UpstreamServiceConfig{},
		HealthStatus: consts.HealthStatusUnhealthy,
	}

	tm := &serviceInfoProviderToolManagerForRepro{
		services: map[string]*tool.ServiceInfo{
			"healthy-service":   srvHealthy,
			"unhealthy-service": srvUnhealthy,
		},
	}

	rm := &mockResourceManagerForRepro{}
	pm := &mockPromptManagerForRepro{
		prompts: []prompt.Prompt{
			&mockPromptForRepro{p: &mcp.Prompt{Name: "healthy-prompt"}, service: "healthy-service"},
			&mockPromptForRepro{p: &mcp.Prompt{Name: "unhealthy-prompt"}, service: "unhealthy-service"},
		},
	}
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// 1. Verify ListPrompts via Middleware
	res, err := server.PromptListFilteringMiddleware(next)(ctx, consts.MethodPromptsList, &mcp.ListPromptsRequest{})
	require.NoError(t, err)
	lRes, ok := res.(*mcp.ListPromptsResult)
	require.True(t, ok)

	foundNames := make(map[string]bool)
	for _, p := range lRes.Prompts {
		foundNames[p.Name] = true
	}

	// Expectation: Unhealthy prompt should NOT be listed
	assert.Contains(t, foundNames, "healthy-prompt")
	assert.NotContains(t, foundNames, "unhealthy-prompt")

	// 2. Verify GetPrompt
	_, err = server.GetPrompt(ctx, &mcp.GetPromptRequest{Params: &mcp.GetPromptParams{
		Name: "unhealthy-prompt",
	}})
	// Expectation: Should fail
	require.Error(t, err)
	assert.Contains(t, err.Error(), consts.HealthStatusUnhealthy)

	// Try getting healthy prompt
	resPrompt, err := server.GetPrompt(ctx, &mcp.GetPromptRequest{Params: &mcp.GetPromptParams{
		Name: "healthy-prompt",
	}})
	require.NoError(t, err)
	require.Empty(t, resPrompt.Messages)
}
