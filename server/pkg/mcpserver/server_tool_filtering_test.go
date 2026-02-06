// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockResourceMgr struct{}
func (m *mockResourceMgr) ListResources() []resource.Resource { return nil }
func (m *mockResourceMgr) GetResource(uri string) (resource.Resource, bool) { return nil, false }
func (m *mockResourceMgr) AddResource(r resource.Resource) {}
func (m *mockResourceMgr) RemoveResource(uri string) {}
func (m *mockResourceMgr) OnListChanged(f func()) {}
func (m *mockResourceMgr) ClearResourcesForService(serviceID string) {}
func (m *mockResourceMgr) Subscribe(ctx context.Context, uri string) error { return nil }


type mockPromptMgr struct{}
func (m *mockPromptMgr) ListPrompts() []prompt.Prompt { return nil }
func (m *mockPromptMgr) GetPrompt(name string) (prompt.Prompt, bool) { return nil, false }
func (m *mockPromptMgr) AddPrompt(p prompt.Prompt) {}
func (m *mockPromptMgr) UpdatePrompt(p prompt.Prompt) {}
func (m *mockPromptMgr) RemovePrompt(name string) {}
func (m *mockPromptMgr) SetMCPServer(s prompt.MCPServerProvider) {}
func (m *mockPromptMgr) ClearPromptsForService(serviceID string) {}


// mockTool implements tool.Tool interface for testing
type mockTool struct {
	definition *v1.Tool
}

func (m *mockTool) Tool() *v1.Tool {
	return m.definition
}

func (m *mockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return "executed", nil
}

func (m *mockTool) MCPTool() *mcp.Tool {
	if m.definition == nil {
		return nil
	}
	return &mcp.Tool{
		Name:        m.definition.GetName(),
		Description: m.definition.GetDescription(),
	}
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestToolFiltering_BugRepro(t *testing.T) {
	// 1. Setup
	ctx := context.Background()
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	tm := tool.NewManager(busProvider)

	// Create a tool "service1.tool1"
	t1 := &v1.Tool{}
	t1.SetName("tool1")
	t1.SetDescription("Tool 1")
	t1.SetServiceId("service1")
	tool1 := &mockTool{definition: t1}

	// Create another tool "service1.tool2" (allowed)
	t2 := &v1.Tool{}
	t2.SetName("tool2")
	t2.SetDescription("Tool 2")
	t2.SetServiceId("service1")
	tool2 := &mockTool{definition: t2}

	// 2. Add tools
	err = tm.AddTool(tool1)
	require.NoError(t, err)
	err = tm.AddTool(tool2)
	require.NoError(t, err)

	// 3. Configure Profiles
	// Profile "user1" enables "service1" but disables "tool1".

	// Construct config objects using Setters because fields are hidden (Opaque API)
	tc1 := &configv1.ToolConfig{}
	tc1.SetDisabled(true)

	psc := &configv1.ProfileServiceConfig{}
	psc.SetEnabled(true)
	psc.SetTools(map[string]*configv1.ToolConfig{"tool1": tc1})

	profileDef := &configv1.ProfileDefinition{}
	profileDef.SetName("user1")
	profileDef.SetServiceConfig(map[string]*configv1.ProfileServiceConfig{"service1": psc})

	tm.SetProfiles([]string{"user1"}, []*configv1.ProfileDefinition{profileDef})

	// 4. Create Server
	authManager := auth.NewManager()

	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	serviceRegistry := serviceregistry.New(factory, tm, &mockPromptMgr{}, &mockResourceMgr{}, authManager)

	server, err := NewServer(ctx, tm, &mockPromptMgr{}, &mockResourceMgr{}, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	// 5. Test "tools/list" with profile
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{}, nil
	}

	ctxUser1 := auth.ContextWithProfileID(ctx, "user1")

	// Invoke middleware directly (accessing private method via same package)
	res, err := server.toolListFilteringMiddleware(next)(ctxUser1, consts.MethodToolsList, &mcp.ListToolsRequest{})
	require.NoError(t, err)

	lRes, ok := res.(*mcp.ListToolsResult)
	require.True(t, ok)

	// 6. Assertions
	foundTools := make(map[string]bool)
	for _, tool := range lRes.Tools {
		foundTools[tool.Name] = true
	}

	// Expect "tool2" (Note: server.go optimization bug also misses namespacing, so it's "tool2" not "service1.tool2")
	// After fix, it should probably be "service1.tool2" if manager handles it correctly.
	// But for repro, we accept either to focus on filtering.
	if _, ok := foundTools["service1.tool2"]; ok {
		assert.Contains(t, foundTools, "service1.tool2", "tool2 should be visible")
	} else {
		assert.Contains(t, foundTools, "tool2", "tool2 should be visible")
	}

	// ⚡ Bolt: This assertion validates the fix. Currently it should FAIL.
	// We check both names just in case.
	assert.NotContains(t, foundTools, "service1.tool1", "tool1 should be disabled and hidden")
	assert.NotContains(t, foundTools, "tool1", "tool1 should be disabled and hidden")
}
