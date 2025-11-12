
package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
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
	mcpServer, err := mcpserver.NewServer(
		ctx,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
	)
	require.NoError(t, err)

	// Add the mock prompt
	promptManager.AddPrompt(&mockPrompt{})

	// Client setup
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := mcpServer.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Test prompts/list
	listResult, err := clientSession.ListPrompts(ctx, &mcp.ListPromptsParams{})
	require.NoError(t, err)
	require.Len(t, listResult.Prompts, 1)
	assert.Equal(t, "test_prompt", listResult.Prompts[0].Name)
	assert.Equal(t, "Test Prompt", listResult.Prompts[0].Title)

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
}
