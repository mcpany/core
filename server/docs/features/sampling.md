# Server-Initiated Sampling

MCP Any supports the MCP "Sampling" capability, which allows tools running on the server to request content generation (sampling) from the connected client (e.g., an LLM). This enables "agentic" workflows where tools can ask clarifying questions, request intermediate thoughts, or generate code snippets using the client's intelligence.

## How it Works

When a tool is executed, it receives a `context.Context`. If the tool is running within an active MCP session (which is the case for tools invoked by an MCP client), the context contains an `MCPSession` (or `Sampler`) interface.

The tool can use this interface to call `CreateMessage`, which sends a `sampling/createMessage` request to the client.

## Usage for Tool Developers

To use sampling in your Go-based tools (implemented via the `Tool` interface):

```go
import (
    "context"
    "github.com/mcpany/core/pkg/tool"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (t *MyTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
    // 1. Retrieve the session from the context
    session, ok := tool.GetSession(ctx)
    if !ok {
        return nil, fmt.Errorf("sampling not available: no active session")
    }

    // 2. Create a sampling request
    params := &mcp.CreateMessageParams{
        Messages: []*mcp.SamplingMessage{
            {
                Role: "user",
                Content: &mcp.TextContent{
                    Text: "Please summarize the following data: ...",
                },
            },
        },
        MaxTokens: 100,
        // Optional: SystemPrompt, Temperature, etc.
    }

    // 3. Send the request to the client
    result, err := session.CreateMessage(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("sampling failed: %w", err)
    }

    // 4. Process the result
    if textContent, ok := result.Content.(*mcp.TextContent); ok {
        return textContent.Text, nil
    }

    return "No text content returned", nil
}
```

## Protocol Support

The server fully handles the underlying JSON-RPC communication for `sampling/createMessage`. It acts as a bridge between the internal tool execution and the connected MCP client.

## Limitations

*   Sampling is only available when the tool is invoked by an MCP client that supports the `sampling` capability.
*   If the tool is invoked via other means (e.g., direct HTTP API call without an MCP session), `tool.GetSession(ctx)` will return `false`. Tools should handle this case gracefully.
