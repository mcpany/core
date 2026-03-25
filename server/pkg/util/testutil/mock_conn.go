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

// NewMockClientConn creates a new mock client conn.
//
// Summary: Creates a new mock client conn.
//
// Parameters: - None.
//   - t (*testing.T): The t.
//
// Returns: - None.
//   - *MockClientConn: The result.
func NewMockClientConn(t *testing.T) *MockClientConn {
	return &MockClientConn{
		t:       t,
		clients: make(map[string]interface{}),
	}
}

// SetClient setClient set client.
//
// Summary: SetClient set client.
//
// Parameters: - None.
//   - method (string): The method.
//   - client (interface{}): The client.
//
// Returns: - None.
//   - None.
func (m *MockClientConn) SetClient(method string, client interface{}) {
	m.clients[method] = client
}

// Invoke invoke invoke.
//
// Summary: Invoke invoke.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - _ (string): Unused parameter.
//   - _ (interface{}): Unused parameter.
//   - _ (interface{}): Unused parameter.
//   - _ (...grpc.CallOption): Unused parameter.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (m *MockClientConn) Invoke(_ context.Context, _ string, _ interface{}, _ interface{}, _ ...grpc.CallOption) error {
	// Not implemented for this mock
	return nil
}

// NewStream creates a new stream.
//
// Summary: Creates a new stream.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - _ (*grpc.StreamDesc): Unused parameter.
//   - method (string): The method.
//   - _ (...grpc.CallOption): Unused parameter.
//
// Returns: - None.
//   - grpc.ClientStream: The result.
//   - error: An error if the operation fails.
func (m *MockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	if client, ok := m.clients[method]; ok {
		return client.(grpc.ClientStream), nil
	}
	return nil, nil
}
