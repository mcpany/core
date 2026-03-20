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
// Summary: Represents a GrpcClient.
type GrpcClient interface {
	// Invoke performs a unary RPC and blocks until the response is received.
	//
	// Parameters:
	//   - ctx: The context for the RPC.
	//   - method: The full gRPC method string (e.g., "/service.Service/Method").
	//   - args: The request message to be sent.
	//   - reply: The response message to be populated.
	//   - opts: gRPC call options.
	// Invoke ...
	//
	// Summary: Executes Invoke operation.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - method: string. A string value.
	//   - args: any. The args parameter.
	//   - reply: any. The reply parameter.
	//   - opts: ...grpc.CallOption. The opts parameter.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

	// NewStream creates a new gRPC stream.
	//
	// Parameters:
	//   - ctx: The context for the stream.
	//   - desc: The stream description.
	//   - method: The full gRPC method string.
	//   - opts: gRPC call options.
	// NewStream ...
	//
	// Summary: Initializes NewStream operation.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - desc: *grpc.StreamDesc. The desc parameter.
	//   - method: string. A string value.
	//   - opts: ...grpc.CallOption. The opts parameter.
	//
	// Returns:
	//   - grpc.ClientStream: The grpc.ClientStream result.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// HTTPClient defines a standard interface for an HTTP client, abstracting the
// underlying implementation. This interface is compatible with the standard
// `*http.Client`.
//
// Summary: Represents a HTTPClient.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// Parameters:
	//   - req: The HTTP request to send.
	// Do ...
	//
	// Summary: Executes Do operation.
	//
	// Parameters:
	//   - req: *http.Request. The request parameters.
	//
	// Returns:
	//   - *http.Response: A pointer to the http.Response result.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	Do(req *http.Request) (*http.Response, error)
}

// MCPClient defines the interface for a client that interacts with an MCP
// service. It provides a standard method for executing tools.
//
// Summary: Represents a MCPClient.
type MCPClient interface {
	// CallTool executes a tool on the MCP service, sending the tool name and
	// inputs and returning the result.
	//
	// Parameters:
	//   - ctx: The context for the call.
	//   - params: The parameters for the tool call, including the tool name and
	//     arguments.
	// CallTool ...
	//
	// Summary: Executes CallTool operation.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - params: *mcp.CallToolParams. The params parameter.
	//
	// Returns:
	//   - *mcp.CallToolResult: A pointer to the mcp.CallToolResult result.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
