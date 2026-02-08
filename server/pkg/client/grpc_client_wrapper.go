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
// Summary: is an interface that represents a gRPC client connection.
type Conn interface {
	grpc.ClientConnInterface
	// Close closes the connection to the server.
	//
	// Summary: closes the connection to the server.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Close() error
	// GetState returns the connectivity.State of the ClientConn.
	//
	// Summary: returns the connectivity.State of the ClientConn.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - connectivity.State: The connectivity.State.
	GetState() connectivity.State
}

// GrpcClientWrapper wraps a `Conn` to adapt it to the.
//
// Summary: wraps a `Conn` to adapt it to the.
type GrpcClientWrapper struct {
	Conn
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewGrpcClientWrapper creates a new GrpcClientWrapper.
//
// Summary: creates a new GrpcClientWrapper.
//
// Parameters:
//   - conn: Conn. The conn.
//   - config: *configv1.UpstreamServiceConfig. The config.
//   - checker: health.Checker. The checker.
//
// Returns:
//   - *GrpcClientWrapper: The *GrpcClientWrapper.
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
// Summary: checks if the underlying gRPC connection is in a usable state.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - bool: The bool.
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

// Close terminates the underlying gRPC connection, releasing any associated.
//
// Summary: terminates the underlying gRPC connection, releasing any associated.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (w *GrpcClientWrapper) Close() error {
	return w.Conn.Close()
}
