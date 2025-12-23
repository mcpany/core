package tool

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/mcpany/core/pkg/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string { return &s }

func TestManager_Pagination_Sorting(t *testing.T) {
	// Setup
	p, err := bus.NewProvider(nil)
	require.NoError(t, err)
	tm := NewManager(p)

	// Add 5 tools with unsorted names
	names := []string{"C", "A", "B", "E", "D"}
	for _, name := range names {
		// Capture variable
		n := name
		tool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return &v1.Tool{
					Name:      stringPtr(n),
					ServiceId: stringPtr("service1"),
				}
			},
		}
		err := tm.AddTool(tool)
		assert.NoError(t, err)
	}

	// 1. Verify ListTools returns sorted tools
	tools := tm.ListTools()
	assert.Len(t, tools, 5)

	expectedOrder := []string{"A", "B", "C", "D", "E"}
	for i, tool := range tools {
		assert.Equal(t, expectedOrder[i], tool.Tool().GetName())
	}
}

func TestManager_Pagination_Paging(t *testing.T) {
	// Setup
	p, err := bus.NewProvider(nil)
	require.NoError(t, err)
	tm := NewManager(p)

	// Add 10 tools A-J
	names := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	for _, name := range names {
		n := name
		tool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return &v1.Tool{
					Name:      stringPtr(n),
					ServiceId: stringPtr("service1"),
				}
			},
		}
		tm.AddTool(tool)
	}

	ctx := context.Background()

	// Page 1: 3 items
	tools, nextCursor, err := tm.ListToolsPaginated(ctx, "", 3, "")
	require.NoError(t, err)
	assert.Len(t, tools, 3)
	assert.Equal(t, "A", tools[0].Tool().GetName())
	assert.Equal(t, "C", tools[2].Tool().GetName())
	assert.NotEmpty(t, nextCursor)

	// Verify cursor encoding (offset=3)
	decoded, _ := base64.StdEncoding.DecodeString(nextCursor)
	assert.Equal(t, "3", string(decoded))

	// Page 2: 3 items
	tools, nextCursor, err = tm.ListToolsPaginated(ctx, nextCursor, 3, "")
	require.NoError(t, err)
	assert.Len(t, tools, 3)
	assert.Equal(t, "D", tools[0].Tool().GetName())
	assert.Equal(t, "F", tools[2].Tool().GetName())
	assert.NotEmpty(t, nextCursor)

	// Page 3: 3 items
	tools, nextCursor, err = tm.ListToolsPaginated(ctx, nextCursor, 3, "")
	require.NoError(t, err)
	assert.Len(t, tools, 3)
	assert.Equal(t, "G", tools[0].Tool().GetName())
	assert.Equal(t, "I", tools[2].Tool().GetName())
	assert.NotEmpty(t, nextCursor)

	// Page 4: 1 item (remaining)
	tools, nextCursor, err = tm.ListToolsPaginated(ctx, nextCursor, 3, "")
	require.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "J", tools[0].Tool().GetName())
	assert.Empty(t, nextCursor, "Should be empty at end")

	// Page 5: 0 items
	// Manually constructing cursor for offset 10
	cursor10 := base64.StdEncoding.EncodeToString([]byte("10"))
	tools, nextCursor, err = tm.ListToolsPaginated(ctx, cursor10, 3, "")
	require.NoError(t, err)
	assert.Empty(t, tools)
	assert.Empty(t, nextCursor)
}

func TestManager_Pagination_ProfileFiltering(t *testing.T) {
	// Setup
	p, err := bus.NewProvider(nil)
	require.NoError(t, err)
	tm := NewManager(p)

	// Add Service Info with Profile
	tm.AddServiceInfo("service_admin", &ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{
			Profiles: []*configv1.Profile{
				{Id: "admin_profile"},
			},
		},
	})
	tm.AddServiceInfo("service_user", &ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{
			Profiles: []*configv1.Profile{
				{Id: "user_profile"},
			},
		},
	})

	// Add tools
	// Admin tool
	tm.AddTool(&MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: stringPtr("AdminTool"), ServiceId: stringPtr("service_admin")}
		},
	})
	// User tool
	tm.AddTool(&MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: stringPtr("UserTool"), ServiceId: stringPtr("service_user")}
		},
	})
	// Public/Legacy tool (no profiles defined in service)
	tm.AddServiceInfo("service_public", &ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{},
	})
	tm.AddTool(&MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: stringPtr("PublicTool"), ServiceId: stringPtr("service_public")}
		},
	})

	ctx := context.Background()

	// 1. Admin Profile: Should see AdminTool and PublicTool?
	// Logic says: if service has profiles, must match. If no profiles, allowed.
	tools, _, err := tm.ListToolsPaginated(ctx, "", 10, "admin_profile")
	require.NoError(t, err)
	// Expect AdminTool and PublicTool. UserTool should be hidden.
	var names []string
	for _, t := range tools {
		names = append(names, t.Tool().GetName())
	}
	assert.ElementsMatch(t, []string{"AdminTool", "PublicTool"}, names)

	// 2. User Profile: Should see UserTool and PublicTool
	tools, _, err = tm.ListToolsPaginated(ctx, "", 10, "user_profile")
	require.NoError(t, err)
	names = nil
	for _, t := range tools {
		names = append(names, t.Tool().GetName())
	}
	assert.ElementsMatch(t, []string{"UserTool", "PublicTool"}, names)

	// 3. Unknown Profile: Should see only PublicTool
	tools, _, err = tm.ListToolsPaginated(ctx, "", 10, "unknown")
	require.NoError(t, err)
	assert.Len(t, tools, 1)
	assert.Equal(t, "PublicTool", tools[0].Tool().GetName())

	// 4. No Profile (empty string): Should see ALL tools?
	// Logic: if profileID == "", filteredTools = allTools
	tools, _, err = tm.ListToolsPaginated(ctx, "", 10, "")
	require.NoError(t, err)
	assert.Len(t, tools, 3)
}
