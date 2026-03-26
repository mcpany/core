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

// ConnectionFactory is responsible for creating new gRPC client connections.
// It can be configured with a custom dialer for testing or special connection
// scenarios.
//
// Summary: Factory component for establishing gRPC connections.
type ConnectionFactory struct {
	dialer func(context.Context, string) (net.Conn, error)
}

// NewConnectionFactory creates and returns a new ConnectionFactory.
//
// Summary: Initializes NewConnectionFactory operation.
//
// Returns:
//   - *ConnectionFactory: The initialized connection factory.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewConnectionFactory() *ConnectionFactory {
	return &ConnectionFactory{}
}

// WithDialer sets a custom dialer function for the ConnectionFactory.
//
// Summary: Executes WithDialer operation.
//
// Parameters:
//   - dialer: func(context.Context, string) (net.Conn, error). The custom dialer function.
//
// Returns:
//   - See return values.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (f *ConnectionFactory) WithDialer(dialer func(context.Context, string) (net.Conn, error)) {
	f.dialer = dialer
}

// NewConnection establishes a new gRPC client connection.
//
// Summary: Initializes NewConnection operation.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - targetAddress: string. The target gRPC server address.
//
// Returns:
//   - *grpc.ClientConn: The established connection.
//   - error: An error if the connection fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
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
