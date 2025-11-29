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

package grpc

import (
	"context"
	"net"
	"strings"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/service"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGrpcPool creates a new connection pool for gRPC clients. It configures the
// pool with a factory function that establishes new gRPC connections with the
// specified address, dialer, and credentials.
//
// minSize is the initial number of connections to create.
// maxSize is the maximum number of connections the pool can hold.
// idleTimeout is the duration after which an idle connection may be closed (not currently implemented).
// address is the target address of the gRPC service.
// dialer is an optional custom dialer for creating network connections.
// creds is the per-RPC credentials to be used for authentication.
// It returns a new gRPC client pool or an error if the pool cannot be created.
func NewGrpcPool(
	minSize, maxSize, idleTimeout int,
	dialer func(context.Context, string) (net.Conn, error),
	creds credentials.PerRPCCredentials,
	config *configv1.UpstreamServiceConfig,
	disableHealthCheck bool,
) (pool.Pool[*client.GrpcClientWrapper], error) {
	factory := func(ctx context.Context) (*client.GrpcClientWrapper, error) {
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		if dialer != nil {
			opts = append(opts, grpc.WithContextDialer(dialer))
		}
		if creds != nil {
			opts = append(opts, grpc.WithPerRPCCredentials(creds))
		}
		if config.GetResilience() != nil && config.GetResilience().GetRetryPolicy() != nil {
			opts = append(opts, grpc.WithUnaryInterceptor(service.UnaryClientInterceptor(config.GetResilience().GetRetryPolicy())))
		}
		addr := strings.TrimPrefix(config.GetGrpcService().GetAddress(), "grpc://")

		conn, err := grpc.NewClient(addr, opts...)
		if err != nil {
			return nil, err
		}
		return client.NewGrpcClientWrapper(conn, config), nil
	}
	return pool.New(factory, minSize, maxSize, idleTimeout, disableHealthCheck)
}
