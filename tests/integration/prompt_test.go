/*
 * Copyright 2025 Author(s) of MCP-XY
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

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/mcpserver"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/serviceregistry"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPrompt struct {
	prompt *mcp.Prompt
}

func (p *testPrompt) Prompt() *mcp.Prompt {
	return p.prompt
}

func (p *testPrompt) Service() string {
	return "test-service"
}

func (p *testPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	var parsedArgs map[string]string
	if err := json.Unmarshal(args, &parsedArgs); err != nil {
		return nil, err
	}
	code := parsedArgs["code"]
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Please review this Python code:\n" + code},
			},
		},
	}, nil
}

func TestPromptIntegration(t *testing.T) {
	t.Skip("Skipping test because the go-sdk v1.0.0 does not support capabilities, which this test asserts.")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Server setup
	busProvider := bus.NewBusProvider()
	poolManager := pool.NewManager()
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(
		factory.NewUpstreamServiceFactory(poolManager),
		toolManager,
		promptManager,
		resourceManager,
		authManager,
	)
	server, err := mcpserver.NewServer(
		ctx,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
	)
	require.NoError(t, err)

	// Add a test prompt
	promptManager.AddPrompt(&testPrompt{
		prompt: &mcp.Prompt{
			Name:        "code_review",
			Title:       "Request Code Review",
			Description: "Asks the LLM to analyze code quality and suggest improvements",
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "code",
					Description: "The code to review",
					Required:    true,
				},
			},
		},
	})

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.1.0"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Verify that the server declared the prompts capability
	assert.NotNil(t, clientSession.InitializeResult().Capabilities.Prompts, "Server did not declare prompts capability")

	// Test prompts/list
	listResult, err := clientSession.ListPrompts(context.Background(), &mcp.ListPromptsParams{})
	require.NoError(t, err)
	assert.Len(t, listResult.Prompts, 1, "Expected one prompt to be available")
	assert.Equal(t, "code_review", listResult.Prompts[0].Name)

	// Test prompts/get
	getResult, err := clientSession.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "code_review",
		Arguments: map[string]string{
			"code": "def hello():\n  print('world')",
		},
	})
	require.NoError(t, err)
	assert.Len(t, getResult.Messages, 1, "Expected one message")
	assert.Equal(t, mcp.Role("user"), getResult.Messages[0].Role)
	textContent, ok := getResult.Messages[0].Content.(*mcp.TextContent)
	assert.True(t, ok, "Expected text content")
	assert.Contains(t, textContent.Text, "def hello():\n  print('world')")
}

func TestPromptLifecycle(t *testing.T) {
	// MCP Server Setup
	serverImpl := &mcp.Implementation{Name: "test-server", Version: "v0.0.1"}
	serverOpts := &mcp.ServerOptions{HasPrompts: true}
	server := mcp.NewServer(serverImpl, serverOpts)

	// MCP Client Setup
	var client *mcp.Client
	var err error

	var wg sync.WaitGroup
	wg.Add(1)
	listChanged := false

	opts := &mcp.ClientOptions{
		PromptListChangedHandler: func(ctx context.Context, req *mcp.PromptListChangedRequest) {
			listChanged = true
			wg.Done()
		},
	}
	client = mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, opts)

	// Use in-memory transports
	t1, t2 := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		_, err := server.Connect(ctx, t1, nil)
		require.NoError(t, err)
	}()

	clientSession, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// 1. Verify that the server correctly declares the `prompts` capability on initialization.
	serverInit := clientSession.InitializeResult()
	assert.NotNil(t, serverInit.Capabilities.Prompts, "Server should declare prompts capability")
	assert.True(t, serverInit.Capabilities.Prompts.ListChanged, "Server should declare ListChanged for prompts")

	// 2. Listing prompts when the list is empty.
	listEmptyResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	assert.Empty(t, listEmptyResult.Prompts, "Prompt list should be empty initially")

	// 3. Adding a prompt and verifying that it appears in the list.
	prompt := &mcp.Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
	}
	server.AddPrompt(prompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "Hello, world!"},
				},
			},
		}, nil
	})

	// Wait for the prompt list to change
	wg.Wait()
	assert.True(t, listChanged, "Prompt list changed notification should be received")

	// 4. Listing prompts again to verify the new prompt is present.
	listResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	require.Len(t, listResult.Prompts, 1, "Prompt list should contain one prompt")
	assert.Equal(t, "test-prompt", listResult.Prompts[0].Name, "Prompt name should match")

	// 5. Getting the content of the added prompt.
	getResult, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "test-prompt",
	})
	require.NoError(t, err)
	assert.NotNil(t, getResult, "GetPrompt result should not be nil")
	require.Len(t, getResult.Messages, 1, "GetPrompt result should contain one message")
	assert.Equal(t, "Hello, world!", getResult.Messages[0].Content.(*mcp.TextContent).Text)

	// 6. Removing the prompt and verifying that the list is empty again.
	wg.Add(1)
	listChanged = false
	server.RemovePrompts("test-prompt")
	wg.Wait()
	assert.True(t, listChanged, "Prompt list changed notification should be received after removal")

	listEmptyResult, err = clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	assert.Empty(t, listEmptyResult.Prompts, "Prompt list should be empty after removal")
}
