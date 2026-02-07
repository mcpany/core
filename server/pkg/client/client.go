// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
//
// Summary: Interface for gRPC client operations.
type GrpcClient interface {
	// Invoke performs a unary RPC and blocks until the response is received.
	//
	// Summary: Executes a unary gRPC call.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the RPC.
	//   - method: string. The full gRPC method string (e.g., "/service.Service/Method").
	//   - args: any. The request message to be sent.
	//   - reply: any. The response message to be populated.
	//   - opts: ...grpc.CallOption. Optional gRPC call options.
	//
	// Returns:
	//   - error: An error if the RPC fails.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

	// NewStream creates a new gRPC stream.
	//
	// Summary: Initiates a new gRPC stream.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the stream.
	//   - desc: *grpc.StreamDesc. The stream description.
	//   - method: string. The full gRPC method string.
	//   - opts: ...grpc.CallOption. Optional gRPC call options.
	//
	// Returns:
	//   - grpc.ClientStream: The client stream interface.
	//   - error: An error if the stream cannot be created.
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// HTTPClient defines a standard interface for an HTTP client, abstracting the
// underlying implementation. This interface is compatible with the standard
// `*http.Client`.
//
// Summary: Interface for HTTP client operations.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// Summary: Executes an HTTP request.
	//
	// Parameters:
	//   - req: *http.Request. The HTTP request to send.
	//
	// Returns:
	//   - *http.Response: The received HTTP response.
	//   - error: An error if the request fails.
	Do(req *http.Request) (*http.Response, error)
}

// MCPClient defines the interface for a client that interacts with an MCP
// service. It provides a standard method for executing tools.
//
// Summary: Interface for MCP client operations.
type MCPClient interface {
	// CallTool executes a tool on the MCP service, sending the tool name and
	// inputs and returning the result.
	//
	// Summary: Calls a tool on an MCP server.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the call.
	//   - params: *mcp.CallToolParams. The parameters for the tool call.
	//
	// Returns:
	//   - *mcp.CallToolResult: The result of the tool execution.
	//   - error: An error if the call fails.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
