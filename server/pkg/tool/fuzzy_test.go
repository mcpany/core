package tool_test

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// simpleMockTool implements tool.Tool interface
type simpleMockTool struct {
	name      string
	serviceID string
}

func (m *simpleMockTool) Tool() *routerv1.Tool {
	t := &routerv1.Tool{}
	t.SetName(m.name)
	t.SetServiceId(m.serviceID)
	return t
}

func (m *simpleMockTool) MCPTool() *mcp.Tool {
	return nil
}

func (m *simpleMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return "executed", nil
}

func (m *simpleMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestManager_ExecuteTool_FuzzyMatching(t *testing.T) {
	// Initialize Manager with nil bus (not needed for this test)
	tm := tool.NewManager(nil)

	// Add some tools
	tools := []string{
		"get_weather",
		"list_files",
		"read_file",
		"create_file",
		"delete_file",
	}

	for _, name := range tools {
		err := tm.AddTool(&simpleMockTool{name: name, serviceID: "test"})
		require.NoError(t, err)
	}

	// Test case: Exact match (should succeed)
	// Note: toolID in manager is serviceID.sanitizedName
	req := &tool.ExecutionRequest{
		ToolName: "test.get_weather",
	}
	_, err := tm.ExecuteTool(context.Background(), req)
	// mockTool returns "executed", so no error.
	require.NoError(t, err)

	// Test case: Typos
	testCases := []struct {
		typo        string
		expectedMsg string
	}{
		{
			typo:        "test.get_wether",
			expectedMsg: "Did you mean: test.get_weather?",
		},
		{
			typo:        "test.list_fils",
			expectedMsg: "Did you mean: test.list_files?",
		},
		{
			typo:        "test.read_fle",
			expectedMsg: "Did you mean: test.read_file?",
		},
		{
			typo: "test.completely_unknown",
			// Should NOT suggest anything if it's too far
			expectedMsg: "unknown tool", // Original error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.typo, func(t *testing.T) {
			req := &tool.ExecutionRequest{
				ToolName: tc.typo,
			}
			_, err := tm.ExecuteTool(context.Background(), req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)

			// Verify it wraps ErrToolNotFound
			assert.True(t, errors.Is(err, tool.ErrToolNotFound), "Error should wrap ErrToolNotFound")
		})
	}
}
