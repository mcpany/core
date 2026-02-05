// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRegister_DynamicResource(t *testing.T) {
	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstream := NewUpstream(nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool"}}}, nil
		},
		listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{}, nil
		},
		listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return &mcp.ListResourcesResult{}, nil
		},
	}

	originalConnect := connectForTesting
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	// Config with dynamic resource
	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("test-tool"),
		CallId: proto.String("call-1"),
	}.Build()

	resDef := configv1.ResourceDefinition_builder{
		Name: proto.String("dynamic-resource"),
		Uri:  proto.String("test://dynamic"),
		Dynamic: configv1.DynamicResource_builder{
			McpCall: configv1.MCPCallDefinition_builder{
				Id: proto.String("call-1"),
			}.Build(),
		}.Build(),
	}.Build()

	mcpService := configv1.McpUpstreamService_builder{
		StdioConnection: configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tools:     []*configv1.ToolDefinition{toolDef},
		Resources: []*configv1.ResourceDefinition{resDef},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:       proto.String("test-service-dynamic"),
		McpService: mcpService,
	}.Build()

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstream := NewUpstream(nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-http"}}}, nil
		},
		listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{}, nil
		},
		listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return &mcp.ListResourcesResult{}, nil
		},
	}

	originalConnect := connectForTesting
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("test-tool-http"),
		CallId: proto.String("call-http-1"),
	}.Build()

	resDef := configv1.ResourceDefinition_builder{
		Name: proto.String("dynamic-resource-http"),
		Uri:  proto.String("test://dynamic-http"),
		Dynamic: configv1.DynamicResource_builder{
			McpCall: configv1.MCPCallDefinition_builder{
				Id: proto.String("call-http-1"),
			}.Build(),
		}.Build(),
	}.Build()

	mcpService := configv1.McpUpstreamService_builder{
		HttpConnection: configv1.McpStreamableHttpConnection_builder{
			HttpAddress: proto.String(server.URL),
		}.Build(),
		Tools:     []*configv1.ToolDefinition{toolDef},
		Resources: []*configv1.ResourceDefinition{resDef},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:       proto.String("test-service-http-dynamic"),
		McpService: mcpService,
	}.Build()

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
	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstream := NewUpstream(nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{
				{Name: "enabled-tool"},
				{Name: "disabled-tool"},
			}}, nil
		},
		listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{
				{Name: "enabled-prompt"},
				{Name: "disabled-prompt"},
			}}, nil
		},
		listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return &mcp.ListResourcesResult{Resources: []*mcp.Resource{
				{URI: "enabled-resource", Name: "enabled-resource"},
				{URI: "disabled-resource", Name: "disabled-resource"},
			}}, nil
		},
	}

	originalConnect := connectForTesting
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	// Configure disabled items
	mcpService := configv1.McpUpstreamService_builder{
		StdioConnection: configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{Name: proto.String("disabled-tool"), Disable: proto.Bool(true)}.Build(),
		},
		Prompts: []*configv1.PromptDefinition{
			configv1.PromptDefinition_builder{Name: proto.String("disabled-prompt"), Disable: proto.Bool(true)}.Build(),
			configv1.PromptDefinition_builder{Name: proto.String("config-disabled-prompt"), Disable: proto.Bool(true)}.Build(), // Also test config-only items
		},
		Resources: []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{Name: proto.String("disabled-resource"), Disable: proto.Bool(true)}.Build(),
			configv1.ResourceDefinition_builder{Name: proto.String("config-disabled-resource"), Disable: proto.Bool(true)}.Build(),
		},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("test-service-disabled"),
		AutoDiscoverTool: proto.Bool(true),
		McpService:       mcpService,
	}.Build()

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		getPromptFunc: func(_ context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
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
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
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
	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstream := NewUpstream(nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{
				{Name: "tool-with-call"},
				{Name: "tool-without-call"},
			}}, nil
		},
		listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{}, nil
		},
		listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return &mcp.ListResourcesResult{}, nil
		},
	}

	originalConnect := connectForTesting
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		defer wg.Done()
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	mcpService := configv1.McpUpstreamService_builder{
		StdioConnection: configv1.McpStdioConnection_builder{
			Command: proto.String("echo"),
		}.Build(),
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{Name: proto.String("tool-with-call"), CallId: proto.String("call-1")}.Build(),
			configv1.ToolDefinition_builder{Name: proto.String("tool-without-call")}.Build(), // No CallId
		},
		Calls: map[string]*configv1.MCPCallDefinition{
			"call-1": configv1.MCPCallDefinition_builder{Id: proto.String("call-1")}.Build(),
		},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:       proto.String("test-service-calls"),
		McpService: mcpService,
	}.Build()

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
