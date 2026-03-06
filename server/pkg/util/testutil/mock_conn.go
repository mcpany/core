// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides testing utilities.
package testutil

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

// MockClientConn is a mock implementation of grpc.ClientConnInterface for testing.
//
// Summary: MockClientConn is a mock implementation of grpc.ClientConnInterface for testing.
//
// Fields:
//   - Contains the configuration and state properties required for MockClientConn functionality.
type MockClientConn struct {
	grpc.ClientConnInterface
	t       *testing.T
	clients map[string]interface{}
}

// NewMockClientConn creates a new mock client connection. Parameters: - t: The testing instance. Returns: - *MockClientConn: A new mock client connection.
//
// Summary: NewMockClientConn creates a new mock client connection. Parameters: - t: The testing instance. Returns: - *MockClientConn: A new mock client connection.
//
// Parameters:
//   - t (*testing.T): The t parameter used in the operation.
//
// Returns:
//   - (*MockClientConn): The resulting MockClientConn object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMockClientConn(t *testing.T) *MockClientConn {
	return &MockClientConn{
		t:       t,
		clients: make(map[string]interface{}),
	}
}

// SetClient sets a mock client for a given type. Parameters: - method: The method to mock. - client: The mock client implementation.
//
// Summary: SetClient sets a mock client for a given type. Parameters: - method: The method to mock. - client: The mock client implementation.
//
// Parameters:
//   - method (string): The method parameter used in the operation.
//   - client (interface{}): The client parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
func (m *MockClientConn) SetClient(method string, client interface{}) {
	m.clients[method] = client
}

// Invoke is a mock implementation of the Invoke method. Parameters: - ctx: The context for the call. - method: The method being invoked. - args: The arguments for the method. - reply: The reply structure to fill. - opts: The call options. Returns: - error: An error if the invocation fails.
//
// Summary: Invoke is a mock implementation of the Invoke method. Parameters: - ctx: The context for the call. - method: The method being invoked. - args: The arguments for the method. - reply: The reply structure to fill. - opts: The call options. Returns: - error: An error if the invocation fails.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - _ (string): The _ parameter used in the operation.
//   - _ (interface{}): The _ parameter used in the operation.
//   - _ (interface{}): The _ parameter used in the operation.
//   - _ (...grpc.CallOption): The _ parameter used in the operation.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (m *MockClientConn) Invoke(_ context.Context, _ string, _ interface{}, _ interface{}, _ ...grpc.CallOption) error {
	// Not implemented for this mock
	return nil
}

// NewStream is a mock implementation of the NewStream method. Parameters: - ctx: The context for the stream. - desc: The stream description. - method: The method being called. - opts: The call options. Returns: - grpc.ClientStream: The client stream. - error: An error if the stream creation fails.
//
// Summary: NewStream is a mock implementation of the NewStream method. Parameters: - ctx: The context for the stream. - desc: The stream description. - method: The method being called. - opts: The call options. Returns: - grpc.ClientStream: The client stream. - error: An error if the stream creation fails.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - _ (*grpc.StreamDesc): The _ parameter used in the operation.
//   - method (string): The method parameter used in the operation.
//   - _ (...grpc.CallOption): The _ parameter used in the operation.
//
// Returns:
//   - (grpc.ClientStream): The resulting grpc.ClientStream object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (m *MockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	if client, ok := m.clients[method]; ok {
		return client.(grpc.ClientStream), nil
	}
	return nil, nil
}
