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
//
// Summary: ConnectionFactory is responsible for creating new gRPC client connections.
type ConnectionFactory struct {
	dialer func(context.Context, string) (net.Conn, error)
}

// NewConnectionFactory creates and returns a new ConnectionFactory with default
//
// Summary: NewConnectionFactory creates and returns a new ConnectionFactory with default
//
// Parameters:
//   - None.
//
// Returns:
//   - *ConnectionFactory: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - *ConnectionFactory: The resulting object or data structure.
//
// Errors:
//   - None.
// WithDialer sets a custom dialer function for the ConnectionFactory. This is
//
// Summary: WithDialer sets a custom dialer function for the ConnectionFactory. This is
//
// Parameters:
//   - dialer (func(context.Context, string) (net.Conn, error)): The textual representation of dialer.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Returns:
//   - None.
//
// NewConnection establishes a new gRPC client connection to the specified
//
// Summary: NewConnection establishes a new gRPC client connection to the specified
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - targetAddress (string): The textual representation of targetaddress.
//
// Returns:
//   - *grpc.ClientConn: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Returns:
//   - *grpc.ClientConn: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
