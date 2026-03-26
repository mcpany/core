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

// Resource ...
// Summary: Resource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.res
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.service
}

// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// Subscribe ...
// Summary: Subscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

type mockResourceManager struct {
	resources []resource.Resource
}

// ListResources ...
// Summary: ListResources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.resources
}

// GetResource ...
// Summary: GetResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	for _, r := range m.resources {
		if r.Resource().URI == uri {
			return r, true
		}
	}
	return nil, false
}

// AddResource ...
// Summary: AddResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.resources = append(m.resources, r)
}

// RemoveResource ...
// Summary: RemoveResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// OnListChanged ...
// Summary: OnListChanged
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ClearResourcesForService ...
// Summary: ClearResourcesForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
}

type mockPrompt struct {
	p       *mcp.Prompt
	service string
}

// Prompt ...
// Summary: Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.p
}

// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.service
}

// Definition ...
// Summary: Definition
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

type mockPromptManager struct {
	prompts []prompt.Prompt
}

// ListPrompts ...
// Summary: ListPrompts
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.prompts
}

// GetPrompt ...
// Summary: GetPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	for _, p := range m.prompts {
		if p.Prompt().Name == name {
			return p, true
		}
	}
	return nil, false
}

// AddPrompt ...
// Summary: AddPrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.prompts = append(m.prompts, p)
}

// UpdatePrompt ...
// Summary: UpdatePrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.prompts = append(m.prompts, p)
}

// RemovePrompt ...
// Summary: RemovePrompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ClearPromptsForService ...
// Summary: ClearPromptsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// reuse smartToolManager concept for GetServiceInfo
type serviceInfoProviderToolManager struct {
	tool.Manager
	services map[string]*tool.ServiceInfo
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	s, ok := m.services[id]
	return s, ok
}

// Stubs
// Stubs
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ExecuteTool ...
// Summary: ExecuteTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// AddMiddleware ...
// Summary: AddMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ClearToolsForService ...
// Summary: ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// IsServiceAllowed ...
// Summary: IsServiceAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// GetAllowedServiceIDs ...
// Summary: GetAllowedServiceIDs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Permissive for testing
	return map[string]bool{
		"global-service":  true,
		"profile-service": true,
		"other-service":   true,
	}, true
}

// TestResourceListFilteringMiddleware ...
// Summary: TestResourceListFilteringMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, nil, busProvider, false)
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

// TestPromptListFilteringMiddleware ...
// Summary: TestPromptListFilteringMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, nil, busProvider, false)
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
