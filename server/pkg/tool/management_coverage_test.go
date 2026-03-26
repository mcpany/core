// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Mock Tool
type mockTool struct {
	toolDef     *v1.Tool
	executeFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool ...
// Summary: Tool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.toolDef
}

// MCPTool ...
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil // Not needed for ExecuteTool tests unless mcp server integration is involved
}

// Execute ...
// Summary: Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return "success", nil
}

// GetCacheConfig ...
// Summary: GetCacheConfig
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

// PreHook Mock
type mockPreHook struct {
	action tool.Action
	err    error
}

// ExecutePre ...
// Summary: ExecutePre
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.action, nil, m.err
}

// PostHook Mock
type mockPostHook struct {
	err error
}

// ExecutePost ...
// Summary: ExecutePost
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return result, m.err
}

// TestManager_ExecuteTool_Coverage ...
// Summary: TestManager_ExecuteTool_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()

	t.Run("Service Unhealthy", func(t *testing.T) {
		manager := tool.NewManager(nil)

		toolDef := v1.Tool_builder{Name: proto.String("unhealthy-tool"), ServiceId: proto.String("unhealthy-svc")}.Build()
		mTool := &mockTool{toolDef: toolDef}

		require.NoError(t, manager.AddTool(mTool))

		info := &tool.ServiceInfo{
			Name:         "unhealthy-svc",
			HealthStatus: "unhealthy",
		}
		manager.AddServiceInfo("unhealthy-svc", info)

		// Use internal ID because ExecuteTool expects it unless nameMap resolves it
		// AddTool adds "unhealthy-svc.unhealthy-tool"
		req := &tool.ExecutionRequest{ToolName: "unhealthy-svc.unhealthy-tool"}
		_, err := manager.ExecuteTool(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service unhealthy-svc is currently unhealthy")
	})

	t.Run("PreHook Deny", func(t *testing.T) {
		manager := tool.NewManager(nil)

		toolDef := v1.Tool_builder{Name: proto.String("deny-tool"), ServiceId: proto.String("deny-svc")}.Build()
		mTool := &mockTool{toolDef: toolDef}

		require.NoError(t, manager.AddTool(mTool))

		info := &tool.ServiceInfo{
			Name:     "deny-svc",
			PreHooks: []tool.PreCallHook{&mockPreHook{action: tool.ActionDeny}},
		}
		manager.AddServiceInfo("deny-svc", info)

		req := &tool.ExecutionRequest{ToolName: "deny-svc.deny-tool"}
		_, err := manager.ExecuteTool(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tool execution denied by hook")
	})

	t.Run("PreHook Error", func(t *testing.T) {
		manager := tool.NewManager(nil)

		toolDef := v1.Tool_builder{Name: proto.String("pre-err-tool"), ServiceId: proto.String("pre-err-svc")}.Build()
		mTool := &mockTool{toolDef: toolDef}

		require.NoError(t, manager.AddTool(mTool))

		info := &tool.ServiceInfo{
			Name:     "pre-err-svc",
			PreHooks: []tool.PreCallHook{&mockPreHook{err: errors.New("pre hook error")}},
		}
		manager.AddServiceInfo("pre-err-svc", info)

		req := &tool.ExecutionRequest{ToolName: "pre-err-svc.pre-err-tool"}
		_, err := manager.ExecuteTool(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pre hook error")
	})

	t.Run("PostHook Error", func(t *testing.T) {
		manager := tool.NewManager(nil)

		toolDef := v1.Tool_builder{Name: proto.String("post-err-tool"), ServiceId: proto.String("post-err-svc")}.Build()
		mTool := &mockTool{toolDef: toolDef}

		require.NoError(t, manager.AddTool(mTool))

		info := &tool.ServiceInfo{
			Name:      "post-err-svc",
			PostHooks: []tool.PostCallHook{&mockPostHook{err: errors.New("post hook error")}},
		}
		manager.AddServiceInfo("post-err-svc", info)

		req := &tool.ExecutionRequest{ToolName: "post-err-svc.post-err-tool"}
		_, err := manager.ExecuteTool(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "post hook error")
	})

	t.Run("Tool Not Found Fuzzy", func(t *testing.T) {
		manager := tool.NewManager(nil)

		toolDef := v1.Tool_builder{Name: proto.String("my-awesome-tool"), ServiceId: proto.String("svc")}.Build()
		mTool := &mockTool{toolDef: toolDef}
		require.NoError(t, manager.AddTool(mTool))

		// "svc.my-awesom-tool" should match "svc.my-awesome-tool"
		req := &tool.ExecutionRequest{ToolName: "svc.my-awesom-tool"}
		_, err := manager.ExecuteTool(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "did you mean \"svc.my-awesome-tool\"")
	})
}

// TestManager_ProfileFiltering_Coverage ...
// Summary: TestManager_ProfileFiltering_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()

	t.Run("AddTool Filtered Out", func(t *testing.T) {
		manager := tool.NewManager(nil)

		// Enable profile "secure"
		profile := configv1.ProfileDefinition_builder{
			Name: proto.String("secure"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"allowed-svc": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
			},
		}.Build()
		manager.SetProfiles([]string{"secure"}, []*configv1.ProfileDefinition{profile})

		// Add tool from blocked service
		toolDef := v1.Tool_builder{Name: proto.String("blocked-tool"), ServiceId: proto.String("blocked-svc")}.Build()
		mTool := &mockTool{toolDef: toolDef}

		err := manager.AddTool(mTool)
		require.NoError(t, err)

		// Check if it was added
		_, found := manager.GetTool("blocked-svc.blocked-tool")
		assert.False(t, found, "Tool should be filtered out by profile")
	})

	t.Run("Tool Matches Profile by Tag", func(t *testing.T) {
		manager := tool.NewManager(nil)

		profile := configv1.ProfileDefinition_builder{
			Name: proto.String("tagged"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"svc": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
			},
			Selector: configv1.ProfileSelector_builder{
				Tags: []string{"public"},
			}.Build(),
		}.Build()
		manager.SetProfiles([]string{"tagged"}, []*configv1.ProfileDefinition{profile})

		toolDef := v1.Tool_builder{
			Name:      proto.String("tagged-tool"),
			ServiceId: proto.String("svc"),
			Tags:      []string{"public"},
		}.Build()
		mTool := &mockTool{toolDef: toolDef}

		err := manager.AddTool(mTool)
		require.NoError(t, err)

		_, found := manager.GetTool("svc.tagged-tool")
		assert.True(t, found, "Tool with tag should be allowed")
	})

	t.Run("Tool Matches Profile by Property", func(t *testing.T) {
		manager := tool.NewManager(nil)

		profile := configv1.ProfileDefinition_builder{
			Name: proto.String("readonly"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"svc": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
			},
			Selector: configv1.ProfileSelector_builder{
				ToolProperties: map[string]string{
					"read_only": "true",
				},
			}.Build(),
		}.Build()
		manager.SetProfiles([]string{"readonly"}, []*configv1.ProfileDefinition{profile})

		toolDef := v1.Tool_builder{
			Name:      proto.String("ro-tool"),
			ServiceId: proto.String("svc"),
			Annotations: v1.ToolAnnotations_builder{
				ReadOnlyHint: proto.Bool(true),
			}.Build(),
		}.Build()
		mTool := &mockTool{toolDef: toolDef}

		err := manager.AddTool(mTool)
		require.NoError(t, err)

		_, found := manager.GetTool("svc.ro-tool")
		assert.True(t, found, "Read-only tool should be allowed")
	})

	t.Run("IsServiceAllowed", func(t *testing.T) {
		manager := tool.NewManager(nil)
		profile := configv1.ProfileDefinition_builder{
			Name: proto.String("test-profile"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"allowed-svc":  configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
				"disabled-svc": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(false)}.Build(),
			},
		}.Build()
		manager.SetProfiles([]string{}, []*configv1.ProfileDefinition{profile})

		assert.True(t, manager.IsServiceAllowed("allowed-svc", "test-profile"))
		assert.False(t, manager.IsServiceAllowed("disabled-svc", "test-profile"))
		assert.False(t, manager.IsServiceAllowed("unknown-svc", "test-profile"))
		assert.False(t, manager.IsServiceAllowed("allowed-svc", "unknown-profile"))
	})

	t.Run("GetAllowedServiceIDs", func(t *testing.T) {
		manager := tool.NewManager(nil)
		profile := configv1.ProfileDefinition_builder{
			Name: proto.String("test-profile"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"allowed-svc": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
			},
		}.Build()
		manager.SetProfiles([]string{}, []*configv1.ProfileDefinition{profile})

		allowed, found := manager.GetAllowedServiceIDs("test-profile")
		require.True(t, found)
		assert.Contains(t, allowed, "allowed-svc")

		_, found = manager.GetAllowedServiceIDs("unknown")
		assert.False(t, found)
	})
}
