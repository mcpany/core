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
//   - conn (Conn): The provided conn data.
//   - config (*configv1.UpstreamServiceConfig): The configuration settings.
//   - checker (health.Checker): The provided checker data.
//
// Returns:
//   - *GrpcClientWrapper: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func NewGrpcClientWrapper(conn Conn, config *configv1.UpstreamServiceConfig, checker health.Checker) *GrpcClientWrapper {
	// If no checker is provided, create a new one (backward compatibility or standalone usage).
	if checker == nil {
		checker = healthChecker.NewChecker(config)
	}
	return &GrpcClientWrapper{
		Conn:    conn,
		config:  config,
		checker: checker,
// IsHealthy checks if the underlying gRPC connection is in a usable state. It returns `true` if the connection's state is not `connectivity.Shutdown`, indicating that it is still active and can be used for new RPCs.
//
// Summary: IsHealthy checks if the underlying gRPC connection is in a usable state. It returns `true` if the connection's state is not `connectivity.Shutdown`, indicating that it is still active and can be used for new RPCs.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (w *GrpcClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.GetState() == connectivity.Shutdown {
		return false
	}
	if w.config.GetGrpcService().GetAddress() == "bufnet" {
		return true
	}
	if w.checker == nil {
// Close terminates the underlying gRPC connection, releasing any associated resources.
//
// Summary: Close terminates the underlying gRPC connection, releasing any associated resources.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (w *GrpcClientWrapper) Close() error {
	return w.Conn.Close()
}
