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
// Summary: Conn is an interface that represents a gRPC client connection.
type Conn interface {
	grpc.ClientConnInterface
	// Close closes the connection to the server.
	//
	// Returns an error if the operation fails.
	Close() error
	// GetState returns the connectivity.State of the ClientConn.
	//
	// Returns the result.
	GetState() connectivity.State
}

// GrpcClientWrapper wraps a `Conn` to adapt it to the
//
// Summary: GrpcClientWrapper wraps a `Conn` to adapt it to the
type GrpcClientWrapper struct {
	Conn
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewGrpcClientWrapper creates a new GrpcClientWrapper. It accepts a shared health checker to avoid creating a new one for every client.
//
// Summary: NewGrpcClientWrapper creates a new GrpcClientWrapper. It accepts a shared health checker to avoid creating a new one for every client.
//
// Parameters:
//   - conn (Conn): The conn parameter.
//   - config (*configv1.UpstreamServiceConfig): The config parameter.
//   - checker (health.Checker): The checker parameter.
//
// Returns:
//   - *GrpcClientWrapper: The *GrpcClientWrapper result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// IsHealthy checks if the underlying gRPC connection is in a usable state. It returns `true` if the connection's state is not `connectivity.Shutdown`, indicating that it is still active and can be used for new RPCs.
//
// Summary: IsHealthy checks if the underlying gRPC connection is in a usable state. It returns `true` if the connection's state is not `connectivity.Shutdown`, indicating that it is still active and can be used for new RPCs.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - bool: The bool result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// Close terminates the underlying gRPC connection, releasing any associated resources.
//
// Summary: Close terminates the underlying gRPC connection, releasing any associated resources.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (w *GrpcClientWrapper) Close() error {
	return w.Conn.Close()
}
