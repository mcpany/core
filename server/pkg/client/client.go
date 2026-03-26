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

// GrpcClient grpcClient represents a grpc client.
//
// Summary: GrpcClient represents a grpc client.
type GrpcClient interface {
	// Invoke performs a unary RPC and blocks until the response is received.
	//
	// Parameters: - None.
	//   - ctx: context.Context. The context for the RPC.
	//   - method: string. The full gRPC method string (e.g., "/service.Service/Method").
	//   - args: any. The request message to be sent.
	//   - reply: any. The response message to be populated.
	//   - opts: ...grpc.CallOption. gRPC call options.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

	// NewStream creates a new gRPC stream.
	//
	// Parameters: - None.
	//   - ctx: context.Context. The context for the stream.
	//   - desc: *grpc.StreamDesc. The stream description.
	//   - method: string. The full gRPC method string.
	//   - opts: ...grpc.CallOption. gRPC call options.
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// HTTPClient hTTPClient represents a http client.
//
// Summary: HTTPClient represents a http client.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// Parameters: - None.
	//   - req: *http.Request. The HTTP request to send.
	Do(req *http.Request) (*http.Response, error)
}

// MCPClient mCPClient represents a mcp client.
//
// Summary: MCPClient represents a mcp client.
type MCPClient interface {
	// CallTool executes a tool on the MCP service, sending the tool name and
	// inputs and returning the result.
	//
	// Parameters: - None.
	//   - ctx: context.Context. The context for the call.
	//   - params: *mcp.CallToolParams. The parameters for the tool call, including the tool name and.
	//     arguments.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
