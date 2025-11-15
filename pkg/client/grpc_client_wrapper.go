/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"

	"github.com/alexliesenfeld/health"
	healthChecker "github.com/mcpany/core/pkg/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Conn is an interface that represents a gRPC client connection.
// It is used to allow for mocking of the gRPC client in tests.
type Conn interface {
	grpc.ClientConnInterface
	Close() error
	GetState() connectivity.State
}

// GrpcClientWrapper wraps a `Conn` to adapt it to the
// `pool.ClosableClient` interface. This allows gRPC clients to be managed by a
// connection pool, which can improve performance by reusing connections.
type GrpcClientWrapper struct {
	Conn
	config *configv1.UpstreamServiceConfig
}

// NewGrpcClientWrapper creates a new GrpcClientWrapper.
//
// Parameters:
//   - conn: The gRPC client connection to be wrapped.
//   - config: The upstream service configuration, used for health checks.
func NewGrpcClientWrapper(conn Conn, config *configv1.UpstreamServiceConfig) *GrpcClientWrapper {
	return &GrpcClientWrapper{
		Conn:   conn,
		config: config,
	}
}

// IsHealthy checks if the underlying gRPC connection is in a usable state.
//
// It returns `true` if the connection's state is not `connectivity.Shutdown`,
// indicating that it is still active and can be used for new RPCs.
func (w *GrpcClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.GetState() == connectivity.Shutdown {
		return false
	}
	if w.config.GetGrpcService().GetAddress() == "bufnet" {
		return true
	}
	checker := healthChecker.NewChecker(w.config)
	if checker == nil {
		return true // No health check configured, assume healthy.
	}
	return checker.Check(ctx).Status == health.StatusUp
}

// Close terminates the underlying gRPC connection, releasing any associated
// resources.
func (w *GrpcClientWrapper) Close() error {
	return w.Conn.Close()
}
