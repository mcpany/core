package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
