/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPPrompt_SimpleMethods(t *testing.T) {
	p := &mcpPrompt{
		mcpPrompt: &mcp.Prompt{Name: "test-prompt"},
		service:   "test-service",
	}

	assert.Equal(t, "test-prompt", p.Prompt().Name)
	assert.Equal(t, "test-service", p.Service())
}

func TestMCPResource_SimpleMethods(t *testing.T) {
	r := &mcpResource{
		mcpResource: &mcp.Resource{URI: "test-uri"},
		service:     "test-service",
	}

	assert.Equal(t, "test-uri", r.Resource().URI)
	assert.Equal(t, "test-service", r.Service())

	err := r.Subscribe(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestAuthenticatedRoundTripper_EdgeCases(t *testing.T) {
	t.Run("nil authenticator", func(t *testing.T) {
		rt := &authenticatedRoundTripper{
			authenticator: nil,
			base:          &mockRoundTripper{},
		}
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		_, err := rt.RoundTrip(req)
		assert.NoError(t, err)
	})

	t.Run("nil base transport", func(t *testing.T) {
		rt := &authenticatedRoundTripper{
			authenticator: nil,
			base:          nil,
		}
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		assert.NotPanics(t, func() {
			rt.RoundTrip(req)
		})
	})
}

func TestStreamableHTTP_NilClient(t *testing.T) {
	transport := &StreamableHTTP{
		Address: "http://example.com",
		Client:  nil,
	}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	assert.NotPanics(t, func() {
		transport.RoundTrip(req)
	})
}

func TestMCPUpstream_Register_DisabledItems(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{
				{Name: "enabled-tool"},
				{Name: "disabled-tool"},
			}}, nil
		},
		listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{
				{Name: "enabled-prompt"},
				{Name: "disabled-prompt"},
			}}, nil
		},
		listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return &mcp.ListResourcesResult{Resources: []*mcp.Resource{
				{URI: "enabled-resource", Name: "enabled-resource"},
				{URI: "disabled-resource", Name: "disabled-resource"},
			}}, nil
		},
	}

	originalConnect := connectForTesting
	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	config := &configv1.UpstreamServiceConfig{}
	config.SetName("test-service-disabled")
	mcpService := &configv1.McpUpstreamService{}
	stdioConnection := &configv1.McpStdioConnection{}
	stdioConnection.SetCommand("echo")
	mcpService.SetStdioConnection(stdioConnection)

	// Disable some items via config
	tool1 := &configv1.ToolDefinition{}
	tool1.SetName("disabled-tool")
	tool1.SetDisable(true)
	mcpService.SetTools([]*configv1.ToolDefinition{tool1})

	prompt1 := &configv1.PromptDefinition{}
	prompt1.SetName("disabled-prompt")
	prompt1.SetDisable(true)
	prompt2 := &configv1.PromptDefinition{}
	prompt2.SetName("disabled-config-prompt")
	prompt2.SetDisable(true)
	mcpService.SetPrompts([]*configv1.PromptDefinition{prompt1, prompt2})

	res1 := &configv1.ResourceDefinition{}
	res1.SetName("disabled-resource")
	res1.SetDisable(true)
	res2 := &configv1.ResourceDefinition{}
	res2.SetName("disabled-config-resource")
	res2.SetDisable(true)
	mcpService.SetResources([]*configv1.ResourceDefinition{res1, res2})

	config.SetMcpService(mcpService)

	serviceID, discoveredTools, discoveredResources, err := upstream.Register(context.Background(), config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	wg.Wait()

	// Verify enabled items are present and disabled ones are absent
	// Tools
	require.Len(t, discoveredTools, 1)
	assert.Equal(t, "enabled-tool", discoveredTools[0].GetName())
	sanitizedEnabledTool, _ := util.SanitizeToolName("enabled-tool")
	_, ok := toolManager.GetTool(serviceID + "." + sanitizedEnabledTool)
	assert.True(t, ok, "enabled-tool should be registered")

	sanitizedDisabledTool, _ := util.SanitizeToolName("disabled-tool")
	_, ok = toolManager.GetTool(serviceID + "." + sanitizedDisabledTool)
	assert.False(t, ok, "disabled-tool should not be registered")

	// Prompts
	_, ok = promptManager.GetPrompt("enabled-prompt")
	assert.True(t, ok, "enabled-prompt should be registered")
	_, ok = promptManager.GetPrompt("disabled-prompt")
	assert.False(t, ok, "disabled-prompt should not be registered")
	_, ok = promptManager.GetPrompt("disabled-config-prompt")
	assert.False(t, ok, "disabled-config-prompt should not be registered")

	// Resources
	require.Len(t, discoveredResources, 1)
	assert.Equal(t, "enabled-resource", discoveredResources[0].GetName())
	_, ok = resourceManager.GetResource("enabled-resource")
	assert.True(t, ok, "enabled-resource should be registered")
	_, ok = resourceManager.GetResource("disabled-resource")
	assert.False(t, ok, "disabled-resource should not be registered")
	_, ok = resourceManager.GetResource("disabled-config-resource")
	assert.False(t, ok, "disabled-config-resource should not be registered")
}

func TestMCPUpstream_Register_DynamicResources(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{
				{Name: "dynamic-tool"},
			}}, nil
		},
		listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{}, nil
		},
		listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return &mcp.ListResourcesResult{}, nil
		},
	}

	originalConnect := connectForTesting
	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	config := &configv1.UpstreamServiceConfig{}
	config.SetName("test-service-dynamic")
	mcpService := &configv1.McpUpstreamService{}
	stdioConnection := &configv1.McpStdioConnection{}
	stdioConnection.SetCommand("echo")
	mcpService.SetStdioConnection(stdioConnection)

	// Define a tool call
	callDef := &configv1.MCPCallDefinition{}
	callDef.SetId("call1")
	mcpService.SetCalls(map[string]*configv1.MCPCallDefinition{
		"call1": callDef,
	})
	// Define the tool that uses the call
	toolDef := &configv1.ToolDefinition{}
	toolDef.SetName("dynamic-tool")
	toolDef.SetCallId("call1")
	mcpService.SetTools([]*configv1.ToolDefinition{toolDef})

	// Define a dynamic resource that uses the tool
	dynamicRes := &configv1.DynamicResource{}
	dynamicCall := &configv1.MCPCallDefinition{}
	dynamicCall.SetId("call1")
	dynamicRes.SetMcpCall(dynamicCall)

	resDef := &configv1.ResourceDefinition{}
	resDef.SetName("dynamic-res")
	resDef.SetDynamic(dynamicRes)

	mcpService.SetResources([]*configv1.ResourceDefinition{resDef})

	config.SetMcpService(mcpService)

	_, _, _, err := upstream.Register(context.Background(), config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	wg.Wait()

	// Verification:
	// Since we don't have an easy way to check if the resource was added without knowing its URI (and it's not static),
	// we rely on the fact that Register succeeded.
	// In a real integration test we would invoke the resource, but here we just check configuration logic.
}

func TestMCPUpstream_Register_MissingConnection(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()

	config := &configv1.UpstreamServiceConfig{}
	config.SetName("test-service-missing-conn")
	mcpService := &configv1.McpUpstreamService{}
	// No stdio or http connection set
	config.SetMcpService(mcpService)

	_, _, _, err := upstream.Register(context.Background(), config, toolManager, promptManager, resourceManager, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires either stdio_connection or http_connection")
}
