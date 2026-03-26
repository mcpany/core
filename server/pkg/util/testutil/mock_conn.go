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
// Summary: Mock implementation of a gRPC client connection for testing.
type MockClientConn struct {
	grpc.ClientConnInterface
	t       *testing.T
	clients map[string]interface{}
}

// NewMockClientConn creates a new mock client connection.
//
// Summary: Initializes NewMockClientConn operation.
//
// Parameters:
//   - t (*testing.T): The testing instance.
//
// Returns:
//   - *MockClientConn: A new mock client connection instance.
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

// SetClient sets a mock client for a given method.
//
// Summary: Updates SetClient operation.
//
// Parameters:
//   - method (string): The method to mock.
//   - client (interface{}): The mock client implementation.
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
// Summary: Executes Invoke operation.
//
// Parameters:
//   - _ (context.Context): Unused context.
//   - _ (string): Unused method.
//   - _ (interface{}): Unused args.
//   - _ (interface{}): Unused reply.
//   - _ (...grpc.CallOption): Unused options.
//
// Returns:
//   - error: Nil.
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
// Summary: Initializes NewStream operation.
//
// Parameters:
//   - _ (context.Context): Unused context.
//   - _ (*grpc.StreamDesc): Unused description.
//   - method (string): The method being called.
//   - _ (...grpc.CallOption): Unused options.
//
// Returns:
//   - grpc.ClientStream: The client stream.
//   - error: Nil.
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
