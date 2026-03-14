// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides testing utilities.
// Summary: MockClientConn is a mock implementation of grpc.ClientConnInterface for testing.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// NewMockClientConn creates a new mock client connection.
//
// Parameters:
//   - t: The testing instance.
//
// Returns:
//   - *MockClientConn: A new mock client connection.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// SetClient sets a mock client for a given type.
//
// Parameters:
//   - method: The method to mock.
//   - client: The mock client implementation.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
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
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
package testutil

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

type MockClientConn struct {
	grpc.ClientConnInterface
	t	*testing.T
	clients	map[string]interface{}
}

func NewMockClientConn(t *testing.T) *MockClientConn {
	return &MockClientConn{
		t:		t,
		clients:	make(map[string]interface{}),
	}
}

func (m *MockClientConn) SetClient(method string, client interface{}) {
	m.clients[method] = client
}

func (m *MockClientConn) Invoke(_ context.Context, _ string, _ interface{}, _ interface{}, _ ...grpc.CallOption) error {
	// Not implemented for this mock
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
	//
	// Errors:
	//   - An error if it fails.
	//
	// Side Effects:
	//   - None.
	return nil
}

func (m *MockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, method string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	if client, ok := m.clients[method]; ok {
		return client.(grpc.ClientStream), nil
	}
	return nil, nil
}
