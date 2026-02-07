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

// --- Mocks ---

type mockResource struct {
	res     *mcp.Resource
	service string
}

func (m *mockResource) Resource() *mcp.Resource {
	return m.res
}

func (m *mockResource) Service() string {
	return m.service
}

func (m *mockResource) Read(_ context.Context) (*mcp.ReadResourceResult, error) {
	return nil, nil
}

func (m *mockResource) Subscribe(_ context.Context) error {
	return nil
}

type mockResourceManager struct {
	resources []resource.Resource
}

func (m *mockResourceManager) ListResources() []resource.Resource {
	return m.resources
}

func (m *mockResourceManager) ListMCPResources() []*mcp.Resource {
	var list []*mcp.Resource
	for _, r := range m.resources {
		list = append(list, r.Resource())
	}
	return list
}

func (m *mockResourceManager) GetResource(uri string) (resource.Resource, bool) {
	for _, r := range m.resources {
		if r.Resource().URI == uri {
			return r, true
		}
	}
	return nil, false
}

func (m *mockResourceManager) AddResource(r resource.Resource) {
	m.resources = append(m.resources, r)
}

func (m *mockResourceManager) RemoveResource(_ string) {}
func (m *mockResourceManager) OnListChanged(f func())    {}
func (m *mockResourceManager) ClearResourcesForService(_ string) {
}

type mockPrompt struct {
	p       *mcp.Prompt
	service string
}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return m.p
}

func (m *mockPrompt) Service() string {
	return m.service
}

func (m *mockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

type mockPromptManager struct {
	prompts []prompt.Prompt
}

func (m *mockPromptManager) ListPrompts() []prompt.Prompt {
	return m.prompts
}

func (m *mockPromptManager) GetPrompt(name string) (prompt.Prompt, bool) {
	for _, p := range m.prompts {
		if p.Prompt().Name == name {
			return p, true
		}
	}
	return nil, false
}

func (m *mockPromptManager) AddPrompt(p prompt.Prompt) {
	m.prompts = append(m.prompts, p)
}

func (m *mockPromptManager) UpdatePrompt(p prompt.Prompt) {
	m.prompts = append(m.prompts, p)
}

func (m *mockPromptManager) RemovePrompt(_ string) {}
func (m *mockPromptManager) SetMCPServer(_ prompt.MCPServerProvider) {}
func (m *mockPromptManager) ClearPromptsForService(_ string) {}


// reuse smartToolManager concept for GetServiceInfo
type serviceInfoProviderToolManager struct {
	tool.Manager
	services map[string]*tool.ServiceInfo
}

func (m *serviceInfoProviderToolManager) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	s, ok := m.services[id]
	return s, ok
}

// Stubs
func (m *serviceInfoProviderToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *serviceInfoProviderToolManager) GetTool(_ string) (tool.Tool, bool)           { return nil, false }
func (m *serviceInfoProviderToolManager) ListTools() []tool.Tool                       { return nil }
func (m *serviceInfoProviderToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *serviceInfoProviderToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *serviceInfoProviderToolManager) SetMCPServer(_ tool.MCPServerProvider)    {}
func (m *serviceInfoProviderToolManager) AddTool(_ tool.Tool) error                { return nil }
func (m *serviceInfoProviderToolManager) ClearToolsForService(_ string)            {}
func (m *serviceInfoProviderToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *serviceInfoProviderToolManager) IsServiceAllowed(_, _ string) bool                       { return true }

func (m *serviceInfoProviderToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	// Permissive for testing
	return map[string]bool{
		"global-service":  true,
		"profile-service": true,
		"other-service":   true,
	}, true
}

func TestResourceListFilteringMiddleware(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Services config
	srvGlobal := &tool.ServiceInfo{Config: configv1.UpstreamServiceConfig_builder{}.Build()}
	srvProfile := &tool.ServiceInfo{Config: configv1.UpstreamServiceConfig_builder{}.Build()}
	srvOther := &tool.ServiceInfo{Config: configv1.UpstreamServiceConfig_builder{}.Build()}

	tm := &serviceInfoProviderToolManager{
		services: map[string]*tool.ServiceInfo{
			"global-service":  srvGlobal,
			"profile-service": srvProfile,
			"other-service":   srvOther,
		},
	}

	rm := &mockResourceManager{
		resources: []resource.Resource{
			&mockResource{res: &mcp.Resource{URI: "global://res"}, service: "global-service"},
			&mockResource{res: &mcp.Resource{URI: "profile://res"}, service: "profile-service"},
			&mockResource{res: &mcp.Resource{URI: "other://res"}, service: "other-service"},
		},
	}
	pm := &mockPromptManager{}
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// 1. No Profile -> Should see ALL resources
	res, err := server.ResourceListFilteringMiddleware(next)(ctx, consts.MethodResourcesList, &mcp.ListResourcesRequest{})
	require.NoError(t, err)
	lRes, ok := res.(*mcp.ListResourcesResult)
	require.True(t, ok)
	assert.Len(t, lRes.Resources, 3)

	// 2. Profile "p1" -> Should see ALL resources
	ctxP1 := auth.ContextWithProfileID(ctx, "p1")
	res, err = server.ResourceListFilteringMiddleware(next)(ctxP1, consts.MethodResourcesList, &mcp.ListResourcesRequest{})
	require.NoError(t, err)
	lRes, ok = res.(*mcp.ListResourcesResult)
	require.True(t, ok)

	foundURIs := make(map[string]bool)
	for _, r := range lRes.Resources {
		foundURIs[r.URI] = true
	}
	assert.Contains(t, foundURIs, "global://res")
	assert.Contains(t, foundURIs, "profile://res")
	assert.Contains(t, foundURIs, "other://res")
	assert.Len(t, lRes.Resources, 3)

	// 3. Profile "p2" -> Should see ALL resources
	ctxP2 := auth.ContextWithProfileID(ctx, "p2")
	res, err = server.ResourceListFilteringMiddleware(next)(ctxP2, consts.MethodResourcesList, &mcp.ListResourcesRequest{})
	require.NoError(t, err)
	lRes, ok = res.(*mcp.ListResourcesResult)
	require.True(t, ok)
	assert.Len(t, lRes.Resources, 3)

	// 4. Other method -> should call next
	res, err = server.ResourceListFilteringMiddleware(next)(ctx, "other/method", nil)
	require.NoError(t, err)
	_, ok = res.(*mcp.CallToolResult)
	require.True(t, ok)
}


func TestPromptListFilteringMiddleware(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Services config
	srvGlobal := &tool.ServiceInfo{Config: configv1.UpstreamServiceConfig_builder{}.Build()}
	srvProfile := &tool.ServiceInfo{Config: configv1.UpstreamServiceConfig_builder{}.Build()}
	srvOther := &tool.ServiceInfo{Config: configv1.UpstreamServiceConfig_builder{}.Build()}

	tm := &serviceInfoProviderToolManager{
		services: map[string]*tool.ServiceInfo{
			"global-service":  srvGlobal,
			"profile-service": srvProfile,
			"other-service":   srvOther,
		},
	}

	rm := &mockResourceManager{}
	pm := &mockPromptManager{
		prompts: []prompt.Prompt{
			&mockPrompt{p: &mcp.Prompt{Name: "global-prompt"}, service: "global-service"},
			&mockPrompt{p: &mcp.Prompt{Name: "profile-prompt"}, service: "profile-service"},
			&mockPrompt{p: &mcp.Prompt{Name: "other-prompt"}, service: "other-service"},
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

	// 1. No Profile -> Should see ALL prompts
	res, err := server.PromptListFilteringMiddleware(next)(ctx, consts.MethodPromptsList, &mcp.ListPromptsRequest{})
	require.NoError(t, err)
	lRes, ok := res.(*mcp.ListPromptsResult)
	require.True(t, ok)
	assert.Len(t, lRes.Prompts, 3)

	// 2. Profile "p1" -> Should see ALL prompts
	ctxP1 := auth.ContextWithProfileID(ctx, "p1")
	res, err = server.PromptListFilteringMiddleware(next)(ctxP1, consts.MethodPromptsList, &mcp.ListPromptsRequest{})
	require.NoError(t, err)
	lRes, ok = res.(*mcp.ListPromptsResult)
	require.True(t, ok)

	foundNames := make(map[string]bool)
	for _, p := range lRes.Prompts {
		foundNames[p.Name] = true
	}
	assert.Contains(t, foundNames, "global-prompt")
	assert.Contains(t, foundNames, "profile-prompt")
	assert.Contains(t, foundNames, "other-prompt")
	assert.Len(t, lRes.Prompts, 3)

	// 3. Profile "p2" -> Should see ALL prompts
	ctxP2 := auth.ContextWithProfileID(ctx, "p2")
	res, err = server.PromptListFilteringMiddleware(next)(ctxP2, consts.MethodPromptsList, &mcp.ListPromptsRequest{})
	require.NoError(t, err)
	lRes, ok = res.(*mcp.ListPromptsResult)
	require.True(t, ok)
	assert.Len(t, lRes.Prompts, 3)

	// 4. Other method -> should call next
	res, err = server.PromptListFilteringMiddleware(next)(ctx, "other/method", nil)
	require.NoError(t, err)
	_, ok = res.(*mcp.CallToolResult)
	require.True(t, ok)
}
