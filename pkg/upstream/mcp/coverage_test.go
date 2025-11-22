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
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	"google.golang.org/protobuf/proto"
)

func TestRegister_DynamicResource(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool"}}}, nil
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

	// Config with dynamic resource
	config := &configv1.UpstreamServiceConfig{}
	config.SetName("test-service-dynamic")
	mcpService := &configv1.McpUpstreamService{}
	stdioConnection := &configv1.McpStdioConnection{}
	stdioConnection.SetCommand("echo")
	mcpService.SetStdioConnection(stdioConnection)

	toolDef := configv1.ToolDefinition_builder{
		Name: proto.String("test-tool"),
		CallId: proto.String("call-1"),
	}.Build()
	mcpService.SetTools([]*configv1.ToolDefinition{toolDef})

	resDef := configv1.ResourceDefinition_builder{
		Name: proto.String("dynamic-resource"),
		Uri: proto.String("test://dynamic"),
		Dynamic: configv1.DynamicResource_builder{
			McpCall: configv1.MCPCallDefinition_builder{
				Id: proto.String("call-1"),
			}.Build(),
		}.Build(),
	}.Build()
	mcpService.SetResources([]*configv1.ResourceDefinition{resDef})

	config.SetMcpService(mcpService)

	serviceID, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	wg.Wait()

	// Verify resource
	r, ok := resourceManager.GetResource("test://dynamic")
	require.True(t, ok)
	_, isDynamic := r.(*resource.DynamicResource)
	assert.True(t, isDynamic)
	assert.Equal(t, serviceID, r.Service())
}

func TestRegister_Http_DynamicResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-http"}}}, nil
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
	config.SetName("test-service-http-dynamic")
	mcpService := &configv1.McpUpstreamService{}
	httpConnection := &configv1.McpStreamableHttpConnection{}
	httpConnection.SetHttpAddress(server.URL)
	mcpService.SetHttpConnection(httpConnection)

	toolDef := configv1.ToolDefinition_builder{
		Name: proto.String("test-tool-http"),
		CallId: proto.String("call-http-1"),
	}.Build()
	mcpService.SetTools([]*configv1.ToolDefinition{toolDef})

	resDef := configv1.ResourceDefinition_builder{
		Name: proto.String("dynamic-resource-http"),
		Uri: proto.String("test://dynamic-http"),
		Dynamic: configv1.DynamicResource_builder{
			McpCall: configv1.MCPCallDefinition_builder{
				Id: proto.String("call-http-1"),
			}.Build(),
		}.Build(),
	}.Build()
	mcpService.SetResources([]*configv1.ResourceDefinition{resDef})

	config.SetMcpService(mcpService)

	serviceID, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	wg.Wait()

	// Verify resource
	r, ok := resourceManager.GetResource("test://dynamic-http")
	require.True(t, ok)
	_, isDynamic := r.(*resource.DynamicResource)
	assert.True(t, isDynamic)
	assert.Equal(t, serviceID, r.Service())
}

func TestRegister_DisabledItems(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()
	ctx := context.Background()

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

	// Configure disabled items
	mcpService.SetTools([]*configv1.ToolDefinition{
		configv1.ToolDefinition_builder{Name: proto.String("disabled-tool"), Disable: proto.Bool(true)}.Build(),
	})
	mcpService.SetPrompts([]*configv1.PromptDefinition{
		configv1.PromptDefinition_builder{Name: proto.String("disabled-prompt"), Disable: proto.Bool(true)}.Build(),
		configv1.PromptDefinition_builder{Name: proto.String("config-disabled-prompt"), Disable: proto.Bool(true)}.Build(), // Also test config-only items
	})
	mcpService.SetResources([]*configv1.ResourceDefinition{
		configv1.ResourceDefinition_builder{Name: proto.String("disabled-resource"), Disable: proto.Bool(true)}.Build(),
		configv1.ResourceDefinition_builder{Name: proto.String("config-disabled-resource"), Disable: proto.Bool(true)}.Build(),
	})

	config.SetMcpService(mcpService)

	serviceID, discoveredTools, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	wg.Wait()

	// Verify tools
	sanitizedEnabled, _ := util.SanitizeToolName("enabled-tool")
	_, ok := toolManager.GetTool(serviceID + "." + sanitizedEnabled)
	assert.True(t, ok, "enabled-tool should be present")

	sanitizedDisabled, _ := util.SanitizeToolName("disabled-tool")
	_, ok = toolManager.GetTool(serviceID + "." + sanitizedDisabled)
	assert.False(t, ok, "disabled-tool should not be present")

	// Verify discovered tools list
	foundDisabled := false
	for _, dt := range discoveredTools {
		if dt.GetName() == "disabled-tool" {
			foundDisabled = true
		}
	}
	assert.False(t, foundDisabled, "disabled-tool should not be in discovered list")

	// Verify prompts
	_, ok = promptManager.GetPrompt("enabled-prompt")
	assert.True(t, ok, "enabled-prompt should be present")

	_, ok = promptManager.GetPrompt("disabled-prompt")
	assert.False(t, ok, "disabled-prompt should not be present")

	_, ok = promptManager.GetPrompt("config-disabled-prompt")
	assert.False(t, ok, "config-disabled-prompt should not be present")

	// Verify resources
	_, ok = resourceManager.GetResource("enabled-resource")
	assert.True(t, ok, "enabled-resource should be present")

	_, ok = resourceManager.GetResource("disabled-resource")
	assert.False(t, ok, "disabled-resource should not be present")
}

func TestPrompt_Get_ComplexArgs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		getPromptFunc: func(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
			// Verify args converted to strings
			assert.Equal(t, "123", params.Arguments["num"])
			assert.Equal(t, "true", params.Arguments["bool"])
			return &mcp.GetPromptResult{}, nil
		},
	}

	conn := &mcpConnection{
		httpAddress: server.URL + "/mcp",
		httpClient:  server.Client(),
		client:      mcp.NewClient(&mcp.Implementation{Name: "test"}, nil),
	}

	originalConnect := connectForTesting
	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	p := &mcpPrompt{
		mcpPrompt:     &mcp.Prompt{Name: "test-prompt"},
		service:       "test-service",
		mcpConnection: conn,
	}

	// Pass complex args
	args := json.RawMessage(`{"num": 123, "bool": true}`)
	_, err := p.Get(context.Background(), args)
	require.NoError(t, err)

	wg.Wait()
}

func TestResource_Subscribe(t *testing.T) {
	r := &mcpResource{
		mcpResource: &mcp.Resource{URI: "test-uri"},
	}
	err := r.Subscribe(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestRegister_CallDefinitionMatching(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	upstream := NewMCPUpstream()
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{
				{Name: "tool-with-call"},
				{Name: "tool-without-call"},
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
	config.SetName("test-service-calls")
	mcpService := &configv1.McpUpstreamService{}
	stdioConnection := &configv1.McpStdioConnection{}
	stdioConnection.SetCommand("echo")
	mcpService.SetStdioConnection(stdioConnection)

	mcpService.SetTools([]*configv1.ToolDefinition{
		configv1.ToolDefinition_builder{Name: proto.String("tool-with-call"), CallId: proto.String("call-1")}.Build(),
		configv1.ToolDefinition_builder{Name: proto.String("tool-without-call")}.Build(), // No CallId
	})

	mcpService.SetCalls(map[string]*configv1.MCPCallDefinition{
		"call-1": configv1.MCPCallDefinition_builder{Id: proto.String("call-1")}.Build(),
	})

	config.SetMcpService(mcpService)

	serviceID, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	wg.Wait()

	// Verify tool-with-call has the call definition
	sanitizedWithCall, _ := util.SanitizeToolName("tool-with-call")
	_, ok := toolManager.GetTool(serviceID + "." + sanitizedWithCall)
	assert.True(t, ok)

	sanitizedWithoutCall, _ := util.SanitizeToolName("tool-without-call")
	_, ok = toolManager.GetTool(serviceID + "." + sanitizedWithoutCall)
	assert.True(t, ok)
}
