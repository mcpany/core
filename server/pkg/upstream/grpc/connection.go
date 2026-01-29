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
type ConnectionFactory struct {
	dialer func(context.Context, string) (net.Conn, error)
}

// NewConnectionFactory creates and returns a new ConnectionFactory with default
// settings.
func NewConnectionFactory() *ConnectionFactory {
	return &ConnectionFactory{}
}

// WithDialer sets a custom dialer function for the ConnectionFactory. This is
// useful for tests that need to mock the network connection.
//
// dialer is the function to be used for creating network connections.
func (f *ConnectionFactory) WithDialer(dialer func(context.Context, string) (net.Conn, error)) {
	f.dialer = dialer
}

// NewConnection establishes a new gRPC client connection to the specified
// target address. It uses insecure credentials by default. If a custom dialer
// has been set, it will be used for the connection.
//
// ctx is the context for the connection attempt.
// targetAddress is the address of the gRPC service to connect to.
// It returns a new *grpc.ClientConn or an error if the connection fails.
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
