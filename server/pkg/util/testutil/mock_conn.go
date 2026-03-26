// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides testing utilities.
package testutil

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

// MockClientConn mockClientConn represents a mock client conn.
//
// Summary: MockClientConn represents a mock client conn.
type MockClientConn struct {
	grpc.ClientConnInterface
	t       *testing.T
	clients map[string]interface{}
}

// NewMockClientConn creates a new mock client connection.
//
// Parameters: - None.
//   - t: The testing instance.
//
// Returns: - None.
//   - *MockClientConn: A new mock client connection.
//
// Summary: Initializes NewMockClientConn operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewMockClientConn(t *testing.T) *MockClientConn {
	return &MockClientConn{
		t:       t,
		clients: make(map[string]interface{}),
	}
}

// SetClient sets a mock client for a given type.
//
// Parameters: - None.
//   - method: The method to mock.
//   - client: The mock client implementation.
//
// Summary: Updates SetClient operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockClientConn) SetClient(method string, client interface{}) {
	m.clients[method] = client
}

// Invoke is a mock implementation of the Invoke method.
//
// Parameters: - None.
//   - ctx: The context for the call.
//   - method: The method being invoked.
//   - args: The arguments for the method.
//   - reply: The reply structure to fill.
//   - opts: The call options.
//
// Returns: - None.
//   - error: An error if the invocation fails.
//
// Summary: Executes Invoke operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockClientConn) Invoke(_ context.Context, _ string, _ interface{}, _ interface{}, _ ...grpc.CallOption) error {
	// Not implemented for this mock
	return nil
}

// NewStream is a mock implementation of the NewStream method.
//
// Parameters: - None.
//   - ctx: The context for the stream.
//   - desc: The stream description.
//   - method: The method being called.
//   - opts: The call options.
//
// Returns: - None.
//   - grpc.ClientStream: The client stream.
//   - error: An error if the stream creation fails.
//
// Summary: Initializes NewStream operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	if client, ok := m.clients[method]; ok {
		return client.(grpc.ClientStream), nil
	}
	return nil, nil
}
