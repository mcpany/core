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

package metrics

import (
	"context"

	"google.golang.org/grpc/stats"
)

// GrpcStatsHandler is a gRPC stats handler that records metrics.
type GrpcStatsHandler struct{}

// TagRPC can be used to tag RPCs with custom information.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC processes RPC stats.
func (h *GrpcStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {}

// TagConn can be used to tag connections with custom information.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes connection stats.
func (h *GrpcStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
		IncrActiveConnections("grpc")
	case *stats.ConnEnd:
		DecrActiveConnections("grpc")
	}
}
