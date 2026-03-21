// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"sync"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUpstream_Register_MergeStrategy(t *testing.T) {
	ctx := context.Background()

	// Helper to create structpb.Struct
	newStruct := func(m map[string]interface{}) *structpb.Struct {
		s, err := structpb.NewStruct(m)
		require.NoError(t, err)
		return s
	}

	t.Run("Override Strategy", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{
					Name: "test-tool",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"original": map[string]interface{}{"type": "string"},
						},
					},
				}}}, nil
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-service-override"),
			AutoDiscoverTool: proto.Bool(true),
			McpService: configv1.McpUpstreamService_builder{
				StdioConnection: configv1.McpStdioConnection_builder{
					Command: proto.String("echo"),
				}.Build(),
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:          proto.String("test-tool"),
						MergeStrategy: configv1.ToolDefinition_MERGE_STRATEGY_OVERRIDE.Enum(),
						InputSchema: newStruct(map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"overridden": map[string]interface{}{"type": "string"},
							},
						}),
					}.Build(),
				},
			}.Build(),
		}.Build()

		serviceID, discoveredTools, _, err := upstream.Register(ctx, config, toolManager, newMockPromptManager(), newMockResourceManager(), false)
		require.NoError(t, err)
		require.Len(t, discoveredTools, 1)

		// Verification
		sanitizedToolName, _ := util.SanitizeToolName("test-tool")
		toolID := serviceID + "." + sanitizedToolName
		tObj, ok := toolManager.GetTool(toolID)
		require.True(t, ok)

		// Check InputSchema - should ONLY have "overridden"
		schema := tObj.Tool().GetInputSchema()
		fields := schema.GetFields()
		props := fields["properties"].GetStructValue().GetFields()
		assert.Contains(t, props, "overridden")
		assert.NotContains(t, props, "original")

		wg.Wait()
	})

	t.Run("Merge Strategy", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{
					Name: "test-tool-merge",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"original": map[string]interface{}{"type": "string"},
							"shared":   map[string]interface{}{"type": "string", "description": "original desc"},
						},
					},
				}}}, nil
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-service-merge"),
			AutoDiscoverTool: proto.Bool(true),
			McpService: configv1.McpUpstreamService_builder{
				StdioConnection: configv1.McpStdioConnection_builder{
					Command: proto.String("echo"),
				}.Build(),
				Tools: []*configv1.ToolDefinition{
					configv1.ToolDefinition_builder{
						Name:          proto.String("test-tool-merge"),
						MergeStrategy: configv1.ToolDefinition_MERGE_STRATEGY_MERGE.Enum(),
						InputSchema: newStruct(map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"new":    map[string]interface{}{"type": "string"},
								"shared": map[string]interface{}{"type": "string", "description": "new desc"},
							},
						}),
					}.Build(),
				},
			}.Build(),
		}.Build()

		serviceID, discoveredTools, _, err := upstream.Register(ctx, config, toolManager, newMockPromptManager(), newMockResourceManager(), false)
		require.NoError(t, err)
		require.Len(t, discoveredTools, 1)

		// Verification
		sanitizedToolName, _ := util.SanitizeToolName("test-tool-merge")
		toolID := serviceID + "." + sanitizedToolName
		tObj, ok := toolManager.GetTool(toolID)
		require.True(t, ok)

		// Check InputSchema - should have "original", "new", and "shared" (with new desc)
		schema := tObj.Tool().GetInputSchema()
		fields := schema.GetFields()
		props := fields["properties"].GetStructValue().GetFields()
		assert.Contains(t, props, "original")
		assert.Contains(t, props, "new")
		assert.Contains(t, props, "shared")

		shared := props["shared"].GetStructValue().GetFields()
		assert.Equal(t, "new desc", shared["description"].GetStringValue())

		wg.Wait()
	})
}
