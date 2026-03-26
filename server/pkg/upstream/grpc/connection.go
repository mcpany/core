// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package grpc provides gRPC upstream integration.
package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectionFactory connectionFactory represents a connection factory.
//
// Summary: ConnectionFactory represents a connection factory.
type ConnectionFactory struct {
	dialer func(context.Context, string) (net.Conn, error)
}

// NewConnectionFactory creates and returns a new ConnectionFactory with default
// settings.
//
// Returns: - None.
//   - *ConnectionFactory: The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewConnectionFactory operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewConnectionFactory() *ConnectionFactory {
	return &ConnectionFactory{}
}

// WithDialer sets a custom dialer function for the ConnectionFactory. This is
// useful for tests that need to mock the network connection.
//
// Parameters: - None.
//   - dialer func(context.Context (string): The parameter.
//   - (string): The parameter.
//
// Returns: - None.
//   - net.Conn: The result.
//   - error): The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes WithDialer operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (f *ConnectionFactory) WithDialer(dialer func(context.Context, string) (net.Conn, error)) {
	f.dialer = dialer
}

// NewConnection establishes a new gRPC client connection to the specified
// target address. It uses insecure credentials by default. If a custom dialer
// has been set, it will be used for the connection.
//
// Parameters: - None.
//   - _ (context.Context): The parameter.
//   - targetAddress (string): The parameter.
//
// Returns: - None.
//   - *grpc.ClientConn: The result.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if ...
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewConnection operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (f *ConnectionFactory) NewConnection(_ context.Context, targetAddress string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if f.dialer != nil {
		opts = append(opts, grpc.WithContextDialer(f.dialer))
	}
	addr := strings.TrimPrefix(targetAddress, "grpc://")

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial target '%s' (address used: '%s'): %w", targetAddress, addr, err)
	}

	return conn, nil
}
