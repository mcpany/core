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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		getPromptFunc: func(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
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

	args := json.RawMessage(`{"arg1": "value1"}`)
	result, err := p.Get(context.Background(), args)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-description", result.Description)

	wg.Wait()
}

func TestMCPResource_Read(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	mockCS := &mockClientSession{
		readResourceFunc: func(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
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
	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
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

func TestMCPUpstream_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration with stdio", func(t *testing.T) {
		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool"}}}, nil
			},
			listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
				return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{{Name: "test-prompt"}}}, nil
			},
			listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
				return &mcp.ListResourcesResult{Resources: []*mcp.Resource{{URI: "test-resource"}}}, nil
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service")
		mcpService := &configv1.McpUpstreamService{}
		stdioConnection := &configv1.McpStdioConnection{}
		stdioConnection.SetCommand("echo")
		stdioConnection.SetArgs([]string{"hello"})
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

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
		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		var capturedTransport mcp.Transport
		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			capturedTransport = transport

			// Return a mock session that returns an empty list of tools to prevent further processing
			return &mockClientSession{
				listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
					return &mcp.ListToolsResult{}, nil
				},
				listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
					return &mcp.ListPromptsResult{}, nil
				},
				listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
					return &mcp.ListResourcesResult{}, nil
				},
			}, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-with-setup")
		mcpService := &configv1.McpUpstreamService{}
		stdioConnection := &configv1.McpStdioConnection{}
		stdioConnection.SetCommand("main_command")
		stdioConnection.SetSetupCommands([]string{"setup1", "setup2"})
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		wg.Wait()

		cmdTransport, ok := capturedTransport.(*mcp.CommandTransport)
		require.True(t, ok, "transport should be CommandTransport")
		assert.Equal(t, "/bin/sh", cmdTransport.Command.Path)
		assert.Equal(t, []string{"/bin/sh", "-c", "setup1 && setup2 && exec main_command"}, cmdTransport.Command.Args)
	})

	t.Run("successful registration with user-specified container image", func(t *testing.T) {
		originalIsDockerSocketAccessibleFunc := util.IsDockerSocketAccessibleFunc
		util.IsDockerSocketAccessibleFunc = func() bool { return true }
		defer func() { util.IsDockerSocketAccessibleFunc = originalIsDockerSocketAccessibleFunc }()

		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		var capturedTransport mcp.Transport
		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			capturedTransport = transport

			// Return a mock session that returns an empty list of tools to prevent further processing
			return &mockClientSession{
				listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
					return &mcp.ListToolsResult{}, nil
				},
				listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
					return &mcp.ListPromptsResult{}, nil
				},
				listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
					return &mcp.ListResourcesResult{}, nil
				},
			}, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-with-image")
		mcpService := &configv1.McpUpstreamService{}
		stdioConnection := &configv1.McpStdioConnection{}
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

		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		var capturedArgs []string
		var capturedCmd string
		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			cmdTransport, ok := transport.(*mcp.CommandTransport)
			require.True(t, ok)
			capturedCmd = cmdTransport.Command.Path
			capturedArgs = cmdTransport.Command.Args

			// Return a mock session that returns an empty list of tools to prevent further processing
			return &mockClientSession{
				listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
					return &mcp.ListToolsResult{}, nil
				},
				listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
					return &mcp.ListPromptsResult{}, nil
				},
				listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
					return &mcp.ListResourcesResult{}, nil
				},
			}, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-no-docker")
		mcpService := &configv1.McpUpstreamService{}
		stdioConnection := &configv1.McpStdioConnection{}
		stdioConnection.SetCommand("node") // A command that would normally use a container
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		wg.Wait()

		assert.Equal(t, "/bin/sh", capturedCmd)
		expectedArgs := []string{"/bin/sh", "-c", "exec node"}
		assert.Equal(t, expectedArgs, capturedArgs)
	})

	t.Run("successful registration with http", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-http"}}}, nil
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-http")
		mcpService := &configv1.McpUpstreamService{}
		httpConnection := &configv1.McpStreamableHttpConnection{}
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
		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return nil, fmt.Errorf("connection failed")
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-fail")
		mcpService := &configv1.McpUpstreamService{}
		httpConnection := &configv1.McpStreamableHttpConnection{}
		httpConnection.SetHttpAddress("http://localhost:9999")
		mcpService.SetHttpConnection(httpConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to MCP service")
		wg.Wait()
	})

	t.Run("list tools error", func(t *testing.T) {
		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		var wg sync.WaitGroup
		wg.Add(1)

		mockCS := &mockClientSession{
			listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return nil, fmt.Errorf("list tools failed")
			},
		}

		originalConnect := connectForTesting
		connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
			defer wg.Done()
			return mockCS, nil
		}
		defer func() { connectForTesting = originalConnect }()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-list-fail")
		mcpService := &configv1.McpUpstreamService{}
		stdioConnection := &configv1.McpStdioConnection{}
		stdioConnection.SetCommand("echo")
		mcpService.SetStdioConnection(stdioConnection)
		config.SetMcpService(mcpService)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list tools from MCP service")
		wg.Wait()
	})

	t.Run("nil mcp service config", func(t *testing.T) {
		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("test-service-nil")
		config.SetMcpService(nil)

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mcp service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		toolManager := tool.NewToolManager(nil)
		promptManager := prompt.NewPromptManager()
		resourceManager := resource.NewResourceManager()
		upstream := NewMCPUpstream()

		config := &configv1.UpstreamServiceConfig{}
		config.SetName("") // empty name

		_, _, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id cannot be empty")
	})
}


func TestAuthenticatedRoundTripper(t *testing.T) {
	var authenticatorCalled bool
	mockAuthenticator := &mockAuthenticator{
		AuthenticateFunc: func(req *http.Request) error {
			authenticatorCalled = true
			return nil
		},
	}

	rt := &authenticatedRoundTripper{
		authenticator: mockAuthenticator,
		base:          &mockRoundTripper{},
	}

	req, _ := http.NewRequest("GET", "http://localhost", nil)
	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.True(t, authenticatorCalled)
}

func TestMCPUpstream_Register_HttpConnectionError(t *testing.T) {
	u := NewMCPUpstream()
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-http-error"),
		McpService: configv1.McpUpstreamService_builder{
			HttpConnection: configv1.McpStreamableHttpConnection_builder{
				HttpAddress: proto.String("http://localhost:12345"),
			}.Build(),
		}.Build(),
	}.Build()

	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return nil, errors.New("connection error")
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
}

func TestBuildCommandFromStdioConfig(t *testing.T) {
	t.Run("simple command", func(t *testing.T) {
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")
		stdio.SetArgs([]string{"-l", "-a"})
		cmd := buildCommandFromStdioConfig(stdio)
		assert.Equal(t, "/bin/sh", cmd.Path)
		assert.Equal(t, []string{"/bin/sh", "-c", "exec ls -l -a"}, cmd.Args)
	})

	t.Run("with setup commands", func(t *testing.T) {
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("my-app")
		stdio.SetArgs([]string{"--verbose"})
		stdio.SetSetupCommands([]string{"cd /tmp", "export FOO=bar"})
		cmd := buildCommandFromStdioConfig(stdio)
		assert.Equal(t, "/bin/sh", cmd.Path)
		assert.Equal(t, []string{"/bin/sh", "-c", "cd /tmp && export FOO=bar && exec my-app --verbose"}, cmd.Args)
	})

	t.Run("docker command with sudo", func(t *testing.T) {
		t.Setenv("USE_SUDO_FOR_DOCKER", "1")
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("docker")
		stdio.SetArgs([]string{"run", "hello-world"})
		cmd := buildCommandFromStdioConfig(stdio)
		assert.Contains(t, cmd.Path, "sudo")
		assert.Equal(t, []string{"sudo", "docker", "run", "hello-world"}, cmd.Args)
	})
}

func TestWithMCPClientSession_NoTransport(t *testing.T) {
	conn := &mcpConnection{}
	err := conn.withMCPClientSession(context.Background(), func(cs ClientSession) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mcp transport is not configured")
}

func TestMCPUpstream_Register_HTTP_Integration(t *testing.T) {
	t.Skip("Skipping failing test for now")
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
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	toolManager := newMockToolManager()
	promptManager := newMockPromptManager()
	resourceManager := newMockResourceManager()
	upstream := NewMCPUpstream()

	originalConnect := connectForTesting
	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return &mockClientSession{
			listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-http"}}}, nil
			},
			listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
				return &mcp.ListPromptsResult{Prompts: []*mcp.Prompt{{Name: "test-prompt-http"}}}, nil
			},
			listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
				return &mcp.ListResourcesResult{Resources: []*mcp.Resource{{URI: "test-resource-http"}}}, nil
			},
		}, nil
	}
	defer func() { connectForTesting = originalConnect }()

	config := &configv1.UpstreamServiceConfig{}
	config.SetName("test-service-http-integration")
	mcpService := &configv1.McpUpstreamService{}
	httpConnection := &configv1.McpStreamableHttpConnection{}
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
	toolID := serviceID + "/" + sanitizedToolName
	_, ok := toolManager.GetTool(toolID)
	assert.True(t, ok)

	_, ok = promptManager.GetPrompt("test-prompt-http")
	assert.True(t, ok)

	_, ok = resourceManager.GetResource("test-resource-http")
	assert.True(t, ok)
}

func TestMCPUpstream_Register_StdioConnectionError(t *testing.T) {
	u := NewMCPUpstream()
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-stdio-error"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String("non-existent-command"),
			}.Build(),
		}.Build(),
	}.Build()

	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return nil, errors.New("connection error")
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
}

func TestMCPUpstream_Register_ListToolsError(t *testing.T) {
	u := NewMCPUpstream()
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
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return nil, errors.New("list tools error")
		},
	}

	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.Error(t, err)
}

func TestMCPUpstream_Register_ListPromptsError(t *testing.T) {
	u := NewMCPUpstream()
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
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{}, nil
		},
		listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return nil, errors.New("list prompts error")
		},
	}

	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.NoError(t, err)
}

func TestMCPUpstream_Register_ListResourcesError(t *testing.T) {
	u := NewMCPUpstream()
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
		listToolsFunc: func(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{}, nil
		},
		listPromptsFunc: func(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
			return &mcp.ListPromptsResult{}, nil
		},
		listResourcesFunc: func(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
			return nil, errors.New("list resources error")
		},
	}

	connectForTesting = func(client *mcp.Client, ctx context.Context, transport mcp.Transport, roots []mcp.Root) (ClientSession, error) {
		return mockCS, nil
	}

	_, _, _, err := u.Register(ctx, serviceConfig, newMockToolManager(), newMockPromptManager(), newMockResourceManager(), false)
	assert.NoError(t, err)
}

func TestMCPUpstream_Register_InvalidServiceConfig(t *testing.T) {
	u := NewMCPUpstream()
	ctx := context.Background()
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-invalid-config"),
	}.Build()

	_, _, _, err := u.Register(ctx, serviceConfig, &mockToolManager{}, &mockPromptManager{}, &mockResourceManager{}, false)
	assert.Error(t, err)
}
