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

// GrpcClient defines a standard interface for a gRPC client, abstracting the.
//
// Summary: defines a standard interface for a gRPC client, abstracting the.
type GrpcClient interface {
	// Invoke performs a unary RPC and blocks until the response is received.
	//
	// Summary: performs a unary RPC and blocks until the response is received.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - method: string. The string.
	//   - args: any. The any.
	//   - reply: any. The any.
	//   - opts: ...grpc.CallOption. The call option.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

	// NewStream creates a new gRPC stream.
	//
	// Summary: creates a new gRPC stream.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - desc: *grpc.StreamDesc. The stream desc.
	//   - method: string. The string.
	//   - opts: ...grpc.CallOption. The call option.
	//
	// Returns:
	//   - grpc.ClientStream: The grpc.ClientStream.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// HTTPClient defines a standard interface for an HTTP client, abstracting the.
//
// Summary: defines a standard interface for an HTTP client, abstracting the.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// Summary: sends an HTTP request and returns an HTTP response.
	//
	// Parameters:
	//   - req: *http.Request. The request object.
	//
	// Returns:
	//   - *http.Response: The *http.Response.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Do(req *http.Request) (*http.Response, error)
}

// MCPClient defines the interface for a client that interacts with an MCP.
//
// Summary: defines the interface for a client that interacts with an MCP.
type MCPClient interface {
	// CallTool executes a tool on the MCP service, sending the tool name and.
	//
	// Summary: executes a tool on the MCP service, sending the tool name and.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - params: *mcp.CallToolParams. The call tool params.
	//
	// Returns:
	//   - *mcp.CallToolResult: The *mcp.CallToolResult.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
