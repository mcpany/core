// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Mock ClientSession for testing
type mockClientSession struct {
	listToolsFunc     func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	listPromptsFunc   func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error)
	listResourcesFunc func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error)
	getPromptFunc     func(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error)
	readResourceFunc  func(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error)
	callToolFunc      func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
	closeFunc         func() error
}

func (m *mockClientSession) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	if m.listToolsFunc != nil {
		return m.listToolsFunc(ctx, params)
	}
	return &mcp.ListToolsResult{}, nil
}

func (m *mockClientSession) ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	if m.listPromptsFunc != nil {
		return m.listPromptsFunc(ctx, params)
	}
	return &mcp.ListPromptsResult{}, nil
}

func (m *mockClientSession) ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	if m.listResourcesFunc != nil {
		return m.listResourcesFunc(ctx, params)
	}
	return &mcp.ListResourcesResult{}, nil
}

func (m *mockClientSession) GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	if m.getPromptFunc != nil {
		return m.getPromptFunc(ctx, params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockClientSession) ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	if m.readResourceFunc != nil {
		return m.readResourceFunc(ctx, params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockClientSession) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockClientSession) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestMCPPrompt_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		getPromptFunc: func(_ context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
			assert.Equal(t, "test-prompt", params.Name)
			assert.Equal(t, "value1", params.Arguments["arg1"])
			return &mcp.GetPromptResult{
				Description: "test-description",
				Messages: []*mcp.PromptMessage{
					{Role: "user", Content: &mcp.TextContent{Text: "Hello"}},
				},
			}, nil
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

	args := json.RawMessage(`{"arg1": "value1"}`)
	result, err := p.Get(context.Background(), args)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-description", result.Description)

	wg.Wait()
}

func TestMCPResource_Read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		readResourceFunc: func(_ context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
			assert.Equal(t, "test-uri", params.URI)
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{URI: "test-uri", Text: "test-content"},
				},
			}, nil
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

	r := &mcpResource{
		mcpResource:   &mcp.Resource{URI: "test-uri"},
		service:       "test-service",
		mcpConnection: conn,
	}

	result, err := r.Read(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
	require.Len(t, result.Contents, 1)
	assert.Equal(t, "test-uri", result.Contents[0].URI)
	assert.Equal(t, "test-content", result.Contents[0].Text)

	wg.Wait()
}

func TestUpstream_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration with stdio", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool"}}}, nil
			},
			listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
				return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{{Name: "test-prompt"}}}, nil
			},
			listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
				return &mcp.ListResourcesResult{Resources: []*mcp.Resource{{URI: "test-resource"}}}, nil
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{
			Name:             proto.String("test-service"),
			AutoDiscoverTool: proto.Bool(true),
			McpService: configv1.McpUpstreamService_builder{
				StdioConnection: configv1.McpStdioConnection_builder{
					Command: proto.String("echo"),
					Args:    []string{"hello"},
				}.Build(),
			}.Build(),
		}.Build()

		serviceID, discoveredTools, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-service")
		assert.Equal(t, expectedKey, serviceID)
		require.Len(t, discoveredTools, 1)
		assert.Equal(t, "test-tool", discoveredTools[0].GetName())

		wg.Wait()

		// Verify registration
		sanitizedToolName, _ := util.SanitizeToolName("test-tool")
		toolID := serviceID + "." + sanitizedToolName
		_, ok := toolManager.GetTool(toolID)
		assert.True(t, ok)
		_, ok = promptManager.GetPrompt("test-prompt")
		assert.True(t, ok)
		_, ok = resourceManager.GetResource("test-resource")
		assert.True(t, ok)
	})

	t.Run("successful registration with stdio and setup commands", func(t *testing.T) {
		// Enable unsafe setup commands for this test
		t.Setenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS", "true")

		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		var capturedTransport mcp.Transport
		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			capturedTransport = transport

			// Return a mock session that returns an empty list of tools to prevent further processing
			return &mockClientSession{
				listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
					return &mcp.ListToolsResult{}, nil
				},
				listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
					return &mcp.ListPromptsResult{}, nil
				},
				listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
					return &mcp.ListResourcesResult{}, nil
				},
			}, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-with-setup")
		mcpService := configv1.McpUpstreamService_builder{}.Build()
		stdioConnection := configv1.McpStdioConnection_builder{}.Build()
		// We use "echo" which is likely to exist.
		stdioConnection.SetCommand("echo")
		stdioConnection.SetSetupCommands([]string{"setup1", "setup2"})
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		wg.Wait()

		cmdTransport, ok := capturedTransport.(*StdioTransport)
		require.True(t, ok, "transport should be StdioTransport")
		assert.Equal(t, "/bin/sh", cmdTransport.Command.Path)
		assert.Equal(t, []string{"/bin/sh", "-c", "setup1 && setup2 && exec echo"}, cmdTransport.Command.Args)
	})

	t.Run("successful registration with user-specified container image", func(t *testing.T) {
		originalIsDockerSocketAccessibleFunc := util.IsDockerSocketAccessibleFunc
		util.IsDockerSocketAccessibleFunc = func() bool { return true }
		defer func() { util.IsDockerSocketAccessibleFunc = originalIsDockerSocketAccessibleFunc }()

		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		var capturedTransport mcp.Transport
		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			capturedTransport = transport

			// Return a mock session that returns an empty list of tools to prevent further processing
			return &mockClientSession{
				listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
					return &mcp.ListToolsResult{}, nil
				},
				listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
					return &mcp.ListPromptsResult{}, nil
				},
				listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
					return &mcp.ListResourcesResult{}, nil
				},
			}, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-with-image")
		mcpService := configv1.McpUpstreamService_builder{}.Build()
		stdioConnection := configv1.McpStdioConnection_builder{}.Build()
		stdioConnection.SetCommand("my-command")
		stdioConnection.SetContainerImage("my-custom-image:latest") // User-specified image
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		wg.Wait()

		dockerTransport, ok := capturedTransport.(*DockerTransport)
		require.True(t, ok, "transport should be DockerTransport")
		assert.Equal(t, "my-custom-image:latest", dockerTransport.StdioConfig.GetContainerImage())
	})

	t.Run("successful registration with docker socket inaccessible", func(t *testing.T) {
		// Override IsDockerSocketAccessible to simulate it being inaccessible.
		originalIsDockerSocketAccessibleFunc := util.IsDockerSocketAccessibleFunc
		util.IsDockerSocketAccessibleFunc = func() bool { return false }
		defer func() { util.IsDockerSocketAccessibleFunc = originalIsDockerSocketAccessibleFunc }()

		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		var capturedArgs []string
		var capturedCmd string
		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			cmdTransport, ok := transport.(*StdioTransport)
			require.True(t, ok)
			capturedCmd = cmdTransport.Command.Path
			capturedArgs = cmdTransport.Command.Args

			// Return a mock session that returns an empty list of tools to prevent further processing
			return &mockClientSession{
				listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
					return &mcp.ListToolsResult{}, nil
				},
				listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
					return &mcp.ListPromptsResult{}, nil
				},
				listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
					return &mcp.ListResourcesResult{}, nil
				},
			}, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-no-docker")
		mcpService := configv1.McpUpstreamService_builder{}.Build()
		stdioConnection := configv1.McpStdioConnection_builder{}.Build()
		stdioConnection.SetCommand("node") // A command that would normally use a container
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		wg.Wait()

		// cmd.Path resolves to the absolute path of "node"
		assert.Contains(t, capturedCmd, "node")
		// Direct execution, so args are just ["node"] (or absolute path)
		assert.Contains(t, capturedArgs[0], "node")
	})

	t.Run("successful registration with http", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-http"}}}, nil
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-http")
		config.SetAutoDiscoverTool(true)
		mcpService := configv1.McpUpstreamService_builder{}.Build()
		httpConnection := configv1.McpStreamableHttpConnection_builder{}.Build()
		httpConnection.SetHttpAddress(server.URL)
		mcpService.SetHttpConnection(httpConnection)
		config.SetMcpService(mcpService)

		serviceID, discoveredTools, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-service-http")
		assert.Equal(t, expectedKey, serviceID)
		require.Len(t, discoveredTools, 1)
		assert.Equal(t, "test-tool-http", discoveredTools[0].GetName())

		wg.Wait()

		sanitizedToolName, _ := util.SanitizeToolName("test-tool-http")
		toolID := serviceID + "." + sanitizedToolName
		_, ok := toolManager.GetTool(toolID)
		assert.True(t, ok)
	})

	t.Run("connection error", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return nil, fmt.Errorf("connection failed")
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-fail")
		mcpService := configv1.McpUpstreamService_builder{}.Build()
		httpConnection := configv1.McpStreamableHttpConnection_builder{}.Build()
		httpConnection.SetHttpAddress("http://127.0.0.1:9999")
		mcpService.SetHttpConnection(httpConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to MCP service")
		wg.Wait()
	})

	t.Run("list tools error", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return nil, fmt.Errorf("list tools failed")
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-list-fail")
		mcpService := configv1.McpUpstreamService_builder{}.Build()
		stdioConnection := configv1.McpStdioConnection_builder{}.Build()
		stdioConnection.SetCommand("echo")
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list tools from MCP service")
		wg.Wait()
	})

	t.Run("nil mcp service config", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("test-service-nil")
		config.SetMcpService(nil)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mcp service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		toolManager := tool.NewManager(nil)
		promptManager := prompt.NewManager()
		resourceManager := resource.NewManager()
		upstream := NewUpstream(nil)

		config := configv1.UpstreamServiceConfig_builder{}.Build()
		config.SetName("") // empty name

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})
}

func TestAuthenticatedRoundTripper(t *testing.T) {
	var authenticatorCalled bool
	mockAuthenticator := &mockAuthenticator{
		AuthenticateFunc: func(_ *http.Request) error {
			authenticatorCalled = true
			return nil
		},
	}

	rt := &authenticatedRoundTripper{
		authenticator: mockAuthenticator,
		base:          &mockRoundTripper{},
	}

	req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	assert.NoError(t, err)
	assert.True(t, authenticatorCalled)
}

func TestUpstream_Register_HttpConnectionError(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-http-error"),
		McpService: configv1.McpUpstreamService_builder{
			HttpConnection: configv1.McpStreamableHttpConnection_builder{
				HttpAddress: proto.String("http://127.0.0.1:12345"),
			}.Build(),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("test-tool"),
					CallId: proto.String("test-call"),
				}.Build(),
			},
			Calls: map[string]*configv1.MCPCallDefinition{
				"test-call": configv1.MCPCallDefinition_builder{
					Id: proto.String("test-call"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return nil, errors.New("connection error")
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
}

func TestBuildCommandFromStdioConfig(t *testing.T) {
	t.Run("simple command", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{}.Build()
		stdio.SetCommand("ls")
		stdio.SetArgs([]string{"-l", "-a"})
		cmd, err := buildCommandFromStdioConfig(context.Background(), stdio, false)
		assert.NoError(t, err)
		assert.Contains(t, cmd.Path, "ls")
		// The first arg in cmd.Args is usually the command name (or path)
		assert.Equal(t, 3, len(cmd.Args))
		assert.Contains(t, cmd.Args[0], "ls")
		assert.Equal(t, "-l", cmd.Args[1])
		assert.Equal(t, "-a", cmd.Args[2])
	})

	t.Run("with setup commands", func(t *testing.T) {
		// Enable unsafe setup commands for this test
		t.Setenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS", "true")

		stdio := configv1.McpStdioConnection_builder{}.Build()
		stdio.SetCommand("echo")
		stdio.SetArgs([]string{"--verbose"})
		stdio.SetSetupCommands([]string{"cd /tmp", "export FOO=bar"})
		cmd, err := buildCommandFromStdioConfig(context.Background(), stdio, false)
		assert.NoError(t, err)
		assert.Equal(t, "/bin/sh", cmd.Path)
		assert.Equal(t, []string{"/bin/sh", "-c", "cd /tmp && export FOO=bar && exec echo --verbose"}, cmd.Args)
	})

	t.Run("with setup commands disabled", func(t *testing.T) {
		// Ensure it's disabled
		t.Setenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS", "")

		stdio := configv1.McpStdioConnection_builder{}.Build()
		stdio.SetCommand("echo")
		stdio.SetSetupCommands([]string{"cd /tmp"})
		cmd, err := buildCommandFromStdioConfig(context.Background(), stdio, false)
		assert.Error(t, err)
		assert.Nil(t, cmd)
		assert.Contains(t, err.Error(), "setup_commands are disabled by default")
	})

	t.Run("docker command with sudo", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{}.Build()
		stdio.SetCommand("docker")
		stdio.SetArgs([]string{"run", "hello-world"})
		cmd, err := buildCommandFromStdioConfig(context.Background(), stdio, true)
		assert.NoError(t, err)
		assert.Contains(t, cmd.Path, "sudo")
		assert.Equal(t, []string{"sudo", "docker", "run", "hello-world"}, cmd.Args)
	})

	t.Run("arguments with shell metacharacters should be passed directly", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{}.Build()
		stdio.SetCommand("echo")
		stdio.SetArgs([]string{"hello; date"})
		cmd, err := buildCommandFromStdioConfig(context.Background(), stdio, false)
		assert.NoError(t, err)
		assert.Contains(t, cmd.Path, "echo")
		assert.Equal(t, 2, len(cmd.Args))
		assert.Equal(t, "hello; date", cmd.Args[1])
	})

	t.Run("attempt complex injection in args", func(t *testing.T) {
		stdio := configv1.McpStdioConnection_builder{}.Build()
		stdio.SetCommand("echo")
		// Attempt to close quote, run command, open quote
		stdio.SetArgs([]string{"foo'; date; echo 'bar"})
		cmd, err := buildCommandFromStdioConfig(context.Background(), stdio, false)
		assert.NoError(t, err)
		// Now we use direct execution, so no shell wrapping
		assert.Contains(t, cmd.Path, "echo")
		// Args should be passed as is, exec will handle them safely
		assert.Equal(t, 2, len(cmd.Args))
		assert.Equal(t, "foo'; date; echo 'bar", cmd.Args[1])
	})
}

func TestWithMCPClientSession_NoTransport(t *testing.T) {
	conn := &mcpConnection{}
	err := conn.withMCPClientSession(context.Background(), func(_ ClientSession) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mcp transport is not configured")
}

func TestUpstream_Register_HTTP_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonrpc.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		var resp jsonrpc.Response
		resp.ID = req.ID
		switch req.Method {
		case "mcp.listTools":
			result := mcp.ListToolsResult{
				Tools: []*mcp.Tool{
					{Name: "test-tool-http"},
				},
			}
			resp.Result, _ = json.Marshal(result)
		case "mcp.listPrompts":
			result := mcp.ListPromptsResult{
				Prompts: []*mcp.Prompt{
					{Name: "test-prompt-http"},
				},
			}
			resp.Result, _ = json.Marshal(result)
		case "mcp.listResources":
			result := mcp.ListResourcesResult{
				Resources: []*mcp.Resource{
					{URI: "test-resource-http"},
				},
			}
			resp.Result, _ = json.Marshal(result)
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	toolManager := newMockToolManager()
	promptManager := newMockPromptManager()
	resourceManager := newMockResourceManager()
	upstream := NewUpstream(nil)

	originalConnect := connectForTesting
	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-http"}}}, nil
			},
			listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
				return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{{Name: "test-prompt-http"}}}, nil
			},
			listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
				return &mcp.ListResourcesResult{Resources: []*mcp.Resource{{URI: "test-resource-http"}}}, nil
			},
		}, nil
	}
	defer func() { connectForTesting = originalConnect }()

	config := configv1.UpstreamServiceConfig_builder{}.Build()
	config.SetName("test-service-http-integration")
	config.SetAutoDiscoverTool(true)
	mcpService := configv1.McpUpstreamService_builder{}.Build()
	httpConnection := configv1.McpStreamableHttpConnection_builder{}.Build()
	httpConnection.SetHttpAddress(server.URL)
	mcpService.SetHttpConnection(httpConnection)
	config.SetMcpService(mcpService)

	serviceID, discoveredTools, discoveredResources, err := upstream.Register(context.Background(), config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	expectedServiceID, _ := util.SanitizeServiceName("test-service-http-integration")
	assert.Equal(t, expectedServiceID, serviceID)
	require.Len(t, discoveredTools, 1)
	assert.Equal(t, "test-tool-http", discoveredTools[0].GetName())
	require.Len(t, discoveredResources, 1)
	assert.Equal(t, "test-resource-http", discoveredResources[0].GetUri())

	sanitizedToolName, _ := util.SanitizeToolName("test-tool-http")
	_, ok := toolManager.GetTool(sanitizedToolName)
	assert.True(t, ok)

	_, ok = promptManager.GetPrompt("test-prompt-http")
	assert.True(t, ok)

	_, ok = resourceManager.GetResource("test-resource-http")
	assert.True(t, ok)
}

func TestUpstream_Register_StdioConnectionError(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-stdio-error"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("non-existent-command"),
			}.Build(),
		}.Build(),
	}.Build()

	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return nil, errors.New("connection error")
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
}

func TestUpstream_Register_ListToolsError(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-list-tools-error"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("echo"),
			}.Build(),
		}.Build(),
	}.Build()

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return nil, errors.New("list tools error")
		},
	}

	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
}

func TestUpstream_Register_ListPromptsError(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-list-prompts-error"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("echo"),
			}.Build(),
		}.Build(),
	}.Build()

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{}, nil
		},
		listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return nil, errors.New("list prompts error")
		},
	}

	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.NoError(t, err)
}

func TestMCPUpstream_Register_ListResourcesError(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-list-resources-error"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("echo"),
			}.Build(),
		}.Build(),
	}.Build()

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{}, nil
		},
		listPromptsFunc: func(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{}, nil
		},
		listResourcesFunc: func(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return nil, errors.New("list resources error")
		},
	}

	connectForTesting = func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.NoError(t, err)
}

func TestMCPUpstream_Register_InvalidServiceConfig(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-invalid-config"),
	}.Build()

	_, _, _, err := u.Register(ctx, serviceConfig, &mockToolManager{}, &mockPromptManager{}, &mockResourceManager{}, false)
	assert.Error(t, err)
}

func TestStreamableHTTP_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "hello")
	}))
	defer server.Close()

	tr := &StreamableHTTP{
		Address: server.URL,
		Client:  server.Client(),
	}

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpstream_Register_InvalidHTTPAddress(t *testing.T) {
	u := NewUpstream(nil)
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-invalid-http-address"),
		McpService: configv1.McpUpstreamService_builder{
			HttpConnection: configv1.McpStreamableHttpConnection_builder{
				HttpAddress: proto.String("file:///etc/passwd"),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mcp http service address scheme")
}
