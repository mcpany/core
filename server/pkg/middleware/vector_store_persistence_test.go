package middleware_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteVectorStore_CallToolResultRoundTrip(t *testing.T) {
	f, err := os.CreateTemp("", "semantic_cache_test_complex_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	store, err := middleware.NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	key := "complex_tool"
	vec := []float32{1.0, 0.0, 0.0}

	// Create a complex CallToolResult
	// Note: TextContent in SDK doesn't have public Type field, it's inferred/internal or used in JSON marshaling?
	// Checking the struct definition (which I can't see but compiler complained about Type field)
	// It seems TextContent struct is: { Text string, ... }
	// And Unmarshaling handles it?

	originalResult := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Hello World",
			},
		},
		IsError: false,
	}

	// Add to store
	err = store.Add(ctx, key, vec, originalResult, 1*time.Minute)
	require.NoError(t, err)

	// Search
	retrievedAny, _, found, err := store.Search(ctx, key, vec)
	require.NoError(t, err)
	require.True(t, found)

	// It comes back as map[string]interface{} because stored as JSON and unmarshaled to any
	retrievedMap, ok := retrievedAny.(map[string]interface{})
	require.True(t, ok, "Expected map[string]interface{}, got %T", retrievedAny)

	// Simulate what server.go:CallTool does: Marshal map back to JSON, then Unmarshal to struct
	jsonBytes, err := json.Marshal(retrievedMap)
	require.NoError(t, err)

	var finalResult mcp.CallToolResult
	err = json.Unmarshal(jsonBytes, &finalResult)
	require.NoError(t, err)

	// Check content
	require.Len(t, finalResult.Content, 1)

	// Issue: The SDK's default UnmarshalJSON might not handle the interface polymorphism automatically
	// unless it has custom logic. Let's see what we got.
	content := finalResult.Content[0]

	// If it failed to determine type, it might be nil or generic map?
	// The SDK defines Content as interface. If standard json unmarshal is used,
	// it might unmarshal into map[string]interface{} if it doesn't know the concrete type.
	// But `Content` is `[]Content`. `Content` is likely `interface{ isContent() }`.
	// Standard JSON unmarshal cannot unmarshal into interface without help.

	// HOWEVER, looking at mcp-go-sdk source (I can't see it directly but I can infer):
	// Usually these SDKs implement UnmarshalJSON.

	// Let's assert what we have.
	// If json unmarshal worked, it should be a *mcp.TextContent
	textContent, ok := content.(*mcp.TextContent)
	if !ok {
		// If it failed to unmarshal to struct, it is likely a map[string]interface{}
		t.Logf("Content type is %T", content)

		// Check if it's *mcp.TextContent, *mcp.ImageContent etc.
		// If generic unmarshal failed to map to concrete type, it might not be in the slice?
		// But CallToolResult has Content []Content.
		// Content is interface.
		// If json.Unmarshal creates something that satisfies Content interface...
		// But map[string]interface{} does not implement MarshalJSON likely.
		// So checking for map is wrong if map doesn't implement it.
		// It implies that Unmarshal MUST return something that implements Content.
		// Which means it must be one of the known types, OR a custom type?
		// If SDK UnmarshalJSON is working, it should be TextContent.
		t.Fail()
	} else {
		assert.Equal(t, "Hello World", textContent.Text)
	}
}
