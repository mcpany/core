// Package client provides the MCP client implementation.

package client

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
)

// GrpcClient defines a standard interface for a gRPC client, abstracting the
// underlying implementation. It provides methods for both unary and streaming
// RPCs and is compatible with the standard `*grpc.ClientConn`.
type GrpcClient interface {
	// Invoke performs a unary RPC and blocks until the response is received.
	//
	// Parameters:
	//   - ctx: The context for the RPC.
	//   - method: The full gRPC method string (e.g., "/service.Service/Method").
	//   - args: The request message to be sent.
	//   - reply: The response message to be populated.
	//   - opts: gRPC call options.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

	// NewStream creates a new gRPC stream.
	//
	// Parameters:
	//   - ctx: The context for the stream.
	//   - desc: The stream description.
	//   - method: The full gRPC method string.
	//   - opts: gRPC call options.
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// HTTPClient defines a standard interface for an HTTP client, abstracting the
// underlying implementation. This interface is compatible with the standard
// `*http.Client`.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// Parameters:
	//   - req: The HTTP request to send.
	Do(req *http.Request) (*http.Response, error)
}

// MCPClient defines the interface for a client that interacts with an MCP
// service. It provides a standard method for executing tools.
type MCPClient interface {
	// CallTool executes a tool on the MCP service, sending the tool name and
	// inputs and returning the result.
	//
	// Parameters:
	//   - ctx: The context for the call.
	//   - params: The parameters for the tool call, including the tool name and
	//     arguments.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
