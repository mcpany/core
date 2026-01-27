// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestIntegrity_ConfigCoverage(t *testing.T) {
	toolDef := &configv1.ToolDefinition{
		Name:        proto.String("config-tool"),
		Description: proto.String("A config tool"),
	}

	// CalculateConfigHash
	hash, err := tool.CalculateConfigHash(toolDef)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// VerifyConfigIntegrity
	err = tool.VerifyConfigIntegrity(toolDef)
	require.NoError(t, err)
}

func TestLocalCommandTool_GetCacheConfig_NilCallDef(t *testing.T) {
	// Test GetCacheConfig when CallDefinition is nil
	toolDef := v1.Tool_builder{
		Name: proto.String("test-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()

	// Create directly struct to bypass NewLocalCommandTool validation or just set fields?
	// NewLocalCommandTool accepts CallDefinition. Pass nil.
	localTool := tool.NewLocalCommandTool(toolDef, service, nil, nil, "id")

	cacheConfig := localTool.GetCacheConfig()
	assert.Nil(t, cacheConfig)
}

func TestOpenAPITool_MCPTool(t *testing.T) {
	t.Run("without service id", func(t *testing.T) {
		toolDef := v1.Tool_builder{
			Name: proto.String("openapi-tool"),
			Description: proto.String("An OpenAPI tool"),
		}.Build()

		var httpCli client.HTTPClient
		openapiTool := tool.NewOpenAPITool(toolDef, httpCli, nil, "GET", "http://example.com", nil, &configv1.OpenAPICallDefinition{})

		mcpTool := openapiTool.MCPTool()
		require.NotNil(t, mcpTool)
		// Name is prefixed with service ID + dot. If empty service ID -> ".name"
		assert.Equal(t, ".openapi-tool", mcpTool.Name)
	})

	t.Run("with service id", func(t *testing.T) {
		toolDef := v1.Tool_builder{
			Name:      proto.String("openapi-tool"),
			ServiceId: proto.String("myservice"),
		}.Build()

		var httpCli client.HTTPClient
		openapiTool := tool.NewOpenAPITool(toolDef, httpCli, nil, "GET", "http://example.com", nil, &configv1.OpenAPICallDefinition{})

		mcpTool := openapiTool.MCPTool()
		require.NotNil(t, mcpTool)
		assert.Equal(t, "myservice.openapi-tool", mcpTool.Name)
	})
}
