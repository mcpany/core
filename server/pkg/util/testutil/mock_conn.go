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
// Summary: Represents a MockClientConn.
type MockClientConn struct {
	grpc.ClientConnInterface
	t       *testing.T
	clients map[string]interface{}
}

// NewMockClientConn creates a new mock client connection.
//
// Summary: Executes the NewMockClientConn operation.
//
// Parameters:
//   - t (*testing.T): The t parameter.
//
// Returns:
//   - *MockClientConn: The returned value.
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

// SetClient sets a mock client for a given type.
//
// Parameters:
//   - method: The method to mock.
//   - client: The mock client implementation.
//
// Summary: Updates SetClient operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
// SetClient sets a mock client for a given type.
//
// Summary: Executes the SetClient operation.
//
// Parameters:
//   - method (string): The method parameter.
//   - client (interface{}): The client parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockClientConn) SetClient(method string, client interface{}) {
	m.clients[method] = client
}

// Invoke is a mock implementation of the Invoke method.
//
// Parameters:
//   - ctx: The context for the call.
//   - method: The method being invoked.
//   - args: The arguments for the method.
//   - reply: The reply structure to fill.
//   - opts: The call options.
//
// Returns:
//   - error: An error if the invocation fails.
//
// Summary: Executes Invoke operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
// Invoke is a mock implementation of the Invoke method.
//
// Summary: Executes the Invoke operation.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (string): The _ parameter.
//   - _ (interface{}): The _ parameter.
//   - _ (interface{}): The _ parameter.
//   - _ (...grpc.CallOption): The _ parameter.
//
// Returns:
//   - error: The returned value.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockClientConn) Invoke(_ context.Context, _ string, _ interface{}, _ interface{}, _ ...grpc.CallOption) error {
	// Not implemented for this mock
	return nil
}

// NewStream is a mock implementation of the NewStream method.
//
// Parameters:
//   - ctx: The context for the stream.
//   - desc: The stream description.
//   - method: The method being called.
//   - opts: The call options.
//
// Returns:
//   - grpc.ClientStream: The client stream.
//   - error: An error if the stream creation fails.
//
// Summary: Initializes NewStream operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
// NewStream is a mock implementation of the NewStream method.
//
// Summary: Executes the NewStream operation.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (*grpc.StreamDesc): The _ parameter.
//   - method (string): The method parameter.
//   - _ (...grpc.CallOption): The _ parameter.
//
// Returns:
//   - grpc.ClientStream: The returned value.
//   - error: The returned value.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	if client, ok := m.clients[method]; ok {
		return client.(grpc.ClientStream), nil
	}
	return nil, nil
}
