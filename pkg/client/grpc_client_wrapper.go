/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// GrpcClientWrapper wraps a `*grpc.ClientConn` to adapt it to the
// `pool.ClosableClient` interface. This allows gRPC clients to be managed by a
// connection pool, which can improve performance by reusing connections.
type GrpcClientWrapper struct {
	*grpc.ClientConn
}

// IsHealthy checks if the underlying gRPC connection is in a usable state.
//
// It returns `true` if the connection's state is not `connectivity.Shutdown`,
// indicating that it is still active and can be used for new RPCs.
func (w *GrpcClientWrapper) IsHealthy() bool {
	return w.GetState() != connectivity.Shutdown
}

// Close terminates the underlying gRPC connection, releasing any associated
// resources.
func (w *GrpcClientWrapper) Close() error {
	return w.ClientConn.Close()
}
