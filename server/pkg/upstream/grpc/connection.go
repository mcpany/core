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
// Summary: Represents a ConnectionFactory.
type ConnectionFactory struct {
	dialer func(context.Context, string) (net.Conn, error)
}

// NewConnectionFactory creates a new connection factory.
//
// Summary: Creates a new connection factory.
//
// Parameters:
//   None.
//
// Returns:
//   - *ConnectionFactory: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewConnectionFactory() *ConnectionFactory {
	return &ConnectionFactory{}
}

// WithDialer withDialer with dialer.
//
// Summary: WithDialer with dialer.
//
// Parameters:
//   - dialer func(context.Context (string): The dialer func(context. context.
//   -  (string): The .
//
// Returns:
//   - net.Conn: The result.
//   - error): The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (f *ConnectionFactory) WithDialer(dialer func(context.Context, string) (net.Conn, error)) {
	f.dialer = dialer
}

// NewConnection creates a new connection.
//
// Summary: Creates a new connection.
//
// Parameters:
//   - _ (context.Context): Unused parameter.
//   - targetAddress (string): The target address.
//
// Returns:
//   - *grpc.ClientConn: The result.
//   - error: An error if the operation fails.
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
