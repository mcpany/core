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

// NewConnectionFactory provides newconnectionfactory functionality.
//
// Summary: NewConnectionFactory.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewConnectionFactory() *ConnectionFactory {
	return &ConnectionFactory{}
}

// WithDialer provides withdialer functionality.
//
// Summary: WithDialer.
//
// Parameters.
//   - dialer: The parameter.
//   - string): The parameter.
//   - error): The parameter.
//
// Returns.
//   - None.
func (f *ConnectionFactory) WithDialer(dialer func(context.Context, string) (net.Conn, error)) {
	f.dialer = dialer
}

// NewConnection provides newconnection functionality.
//
// Summary: NewConnection.
//
// Parameters.
//   - _: The parameter.
//   - targetAddress: The parameter.
//   - error: The parameter.
//
// Returns.
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
