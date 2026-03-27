// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Conn is an interface that represents a gRPC client connection.
// It is used to allow for mocking of the gRPC client in tests.
//
// Summary: Represents a Conn.
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
// `pool.ClosableClient` interface. This allows gRPC clients to be managed by a
// connection pool, which can improve performance by reusing connections.
//
// Summary: Represents a GrpcClientWrapper.
type GrpcClientWrapper struct {
	Conn
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewGrpcClientWrapper provides newgrpcclientwrapper functionality.
//
// Summary: NewGrpcClientWrapper.
//
// Parameters.
//   - conn: The parameter.
//   - config: The parameter.
//   - checker: The parameter.
//
// Returns.
//   - result: The result.
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

// IsHealthy provides ishealthy functionality.
//
// Summary: IsHealthy.
//
// Parameters.
//   - ctx: The parameter.
//
// Returns.
//   - result: The result.
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

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (w *GrpcClientWrapper) Close() error {
	return w.Conn.Close()
}
