package integration

import (
	"context"
	"encoding/json"
	"sync"
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

func (p *testPrompt) Get(_ context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
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
	server, err := mcpserver.NewServer(
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
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

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
		PromptListChangedHandler: func(_ context.Context, _ *mcp.PromptListChangedRequest) {
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
	defer func() { _ = clientSession.Close() }()

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
	server.AddPrompt(prompt, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
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
