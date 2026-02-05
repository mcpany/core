// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// mockPrompt is a mock implementation of the prompt.Prompt interface for testing.
type mockPrompt struct{}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{
		Name:        "test_prompt",
		Title:       "Test Prompt",
		Description: "A prompt for testing.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "test_arg",
				Description: "A test argument.",
				Required:    true,
			},
		},
	}
}

func (m *mockPrompt) Service() string {
	return "test_service"
}

func (m *mockPrompt) Get(
	_ context.Context,
	_ json.RawMessage,
) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "test content"},
			},
		},
	}, nil
}

func TestPromptsEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Server setup
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	poolManager := pool.NewManager()
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(
		factory.NewUpstreamServiceFactory(poolManager, nil),
		toolManager,
		promptManager,
		resourceManager,
		authManager,
	)
	mcpServer, err := mcpserver.NewServer(
		ctx,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		false,
	)
	require.NoError(t, err)

	// Add the mock prompt
	promptManager.AddPrompt(&mockPrompt{})

	// Add a templated prompt
	role := configv1.PromptMessage_USER
	definition := configv1.PromptDefinition_builder{
		Name: proto.String("templated-prompt"),
		Messages: []*configv1.PromptMessage{
			configv1.PromptMessage_builder{
				Role: &role,
				Text: configv1.TextContent_builder{
					Text: proto.String("Hello, {{name}}!"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	promptManager.AddPrompt(prompt.NewTemplatedPrompt(definition, "test-service"))

	// Client setup
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := mcpServer.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Test prompts/list
	listResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	require.Len(t, listResult.Prompts, 2)

	// Check for the mock prompt
	foundMockPrompt := false
	for _, p := range listResult.Prompts {
		if p.Name == "test_prompt" {
			assert.Equal(t, "Test Prompt", p.Title)
			foundMockPrompt = true
			break
		}
	}
	assert.True(t, foundMockPrompt, "mock prompt not found")

	// Check for the templated prompt
	foundTemplatedPrompt := false
	for _, p := range listResult.Prompts {
		if p.Name == "test-service.templated-prompt" {
			foundTemplatedPrompt = true
			break
		}
	}
	assert.True(t, foundTemplatedPrompt, "templated prompt not found")

	// Test prompts/get
	getResult, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "test_prompt",
		Arguments: map[string]string{
			"test_arg": "test_value",
		},
	})
	require.NoError(t, err)
	require.Len(t, getResult.Messages, 1)
	textContent, ok := getResult.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "test content", textContent.Text)

	// Test templated prompts/get
	getResult, err = clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "test-service.templated-prompt",
		Arguments: map[string]string{
			"name": "world",
		},
	})
	require.NoError(t, err)
	require.Len(t, getResult.Messages, 1)
	textContent, ok = getResult.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Hello, world!", textContent.Text)
}
