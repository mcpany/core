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
// Summary: is a mock implementation of grpc.ClientConnInterface for testing.
type MockClientConn struct {
	grpc.ClientConnInterface
	t       *testing.T
	clients map[string]interface{}
}

// NewMockClientConn creates a new mock client connection.
//
// Summary: creates a new mock client connection.
//
// Parameters:
//   - t: *testing.T. The t.
//
// Returns:
//   - *MockClientConn: The *MockClientConn.
func NewMockClientConn(t *testing.T) *MockClientConn {
	return &MockClientConn{
		t:       t,
		clients: make(map[string]interface{}),
	}
}

// SetClient sets a mock client for a given type.
//
// Summary: sets a mock client for a given type.
//
// Parameters:
//   - method: string. The method.
//   - client: interface{}. The client.
//
// Returns:
//   None.
func (m *MockClientConn) SetClient(method string, client interface{}) {
	m.clients[method] = client
}

// Invoke is a mock implementation of the Invoke method.
//
// Summary: is a mock implementation of the Invoke method.
//
// Parameters:
//   - _: context.Context. The _.
//   - _: string. The _.
//   - _: interface{}. The _.
//   - _: interface{}. The _.
//   - _: ...grpc.CallOption. The _.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *MockClientConn) Invoke(_ context.Context, _ string, _ interface{}, _ interface{}, _ ...grpc.CallOption) error {
	// Not implemented for this mock
	return nil
}

// NewStream is a mock implementation of the NewStream method.
//
// Summary: is a mock implementation of the NewStream method.
//
// Parameters:
//   - _: context.Context. The _.
//   - _: *grpc.StreamDesc. The _.
//   - method: string. The method.
//   - _: ...grpc.CallOption. The _.
//
// Returns:
//   - grpc.ClientStream: The grpc.ClientStream.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *MockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	if client, ok := m.clients[method]; ok {
		return client.(grpc.ClientStream), nil
	}
	return nil, nil
}
