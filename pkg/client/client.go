/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
)

// GrpcClient defines a standard interface for a gRPC client, abstracting the
// underlying implementation. It provides methods for both unary and streaming
// RPCs. This interface is compatible with *grpc.ClientConn.
type GrpcClient interface {
	// Invoke performs a unary RPC and blocks until the response is received.
	//
	// ctx is the context for the RPC.
	// method is the full gRPC method string.
	// args is the request message to be sent.
	// reply is the response message to be populated.
	// opts are the gRPC call options.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error

	// NewStream creates a new gRPC stream.
	//
	// ctx is the context for the stream.
	// desc is the stream description.
	// method is the full gRPC method string.
	// opts are the gRPC call options.
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// HttpClient defines a standard interface for an HTTP client. This interface is
// compatible with *http.Client.
type HttpClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// req is the HTTP request to send.
	Do(req *http.Request) (*http.Response, error)
}

// MCPClient defines the interface for a client that interacts with an MCP
// service. It provides a method for calling tools.
type MCPClient interface {
	// CallTool executes a tool on the MCP service.
	//
	// ctx is the context for the call.
	// params contains the parameters for the tool call, including the tool name
	// and inputs.
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}
