// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	"github.com/alexliesenfeld/health"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Conn is an interface that represents a gRPC client connection.
//
// Summary: Interface for gRPC client connection.
//
// It is used to allow for mocking of the gRPC client in tests.
type Conn interface {
	grpc.ClientConnInterface
	// Close closes the connection to the server.
	//
	// Summary: Closes the connection.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Close() error
	// GetState returns the connectivity.State of the ClientConn.
	//
	// Summary: Returns connection state.
	//
	// Returns:
	//   - connectivity.State: The current state.
	GetState() connectivity.State
}

// GrpcClientWrapper wraps a `Conn` to adapt it to the
// `pool.ClosableClient` interface.
//
// Summary: Wrapper for gRPC connection to support pooling.
//
// This allows gRPC clients to be managed by a
// connection pool, which can improve performance by reusing connections.
//
// Fields:
//   - Conn: Conn. The underlying gRPC connection.
//   - config: *configv1.UpstreamServiceConfig. The service configuration.
//   - checker: health.Checker. The health checker.
type GrpcClientWrapper struct {
	Conn
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewGrpcClientWrapper creates a new GrpcClientWrapper.
//
// Summary: Creates a new GrpcClientWrapper.
//
// It accepts a shared health checker to avoid creating a new one for every client.
//
// Parameters:
//   - conn: Conn. The underlying gRPC connection.
//   - config: *configv1.UpstreamServiceConfig. The service configuration.
//   - checker: health.Checker. The health checker.
//
// Returns:
//   - *GrpcClientWrapper: The new wrapper instance.
func NewGrpcClientWrapper(conn Conn, config *configv1.UpstreamServiceConfig, checker health.Checker) *GrpcClientWrapper {
	// If no checker is provided, create a new one (backward compatibility or standalone usage).
	if checker == nil {
		checker = healthChecker.NewChecker(config)
	}
	return &GrpcClientWrapper{
		Conn:    conn,
		config:  config,
		checker: checker,
	}
}

// IsHealthy checks if the underlying gRPC connection is in a usable state.
//
// Summary: Checks connection health.
//
// It returns `true` if the connection's state is not `connectivity.Shutdown`,
// indicating that it is still active and can be used for new RPCs.
//
// Parameters:
//   - ctx: context.Context. The context for the check.
//
// Returns:
//   - bool: True if healthy.
func (w *GrpcClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.GetState() == connectivity.Shutdown {
		return false
	}
	if w.config.GetGrpcService().GetAddress() == "bufnet" {
		return true
	}
	if w.checker == nil {
		return true // No health check configured, assume healthy.
	}
	return w.checker.Check(ctx).Status == health.StatusUp
}

// Close terminates the underlying gRPC connection, releasing any associated
// resources.
//
// Summary: Closes the connection.
//
// Returns:
//   - error: An error if closing fails.
func (w *GrpcClientWrapper) Close() error {
	return w.Conn.Close()
}
