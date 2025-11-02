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

package middleware

import (
	"context"
	"time"

	"github.com/mcpany/core/pkg/metrics"
	"google.golang.org/grpc"
)

// GRPCMetricsInterceptor is a gRPC interceptor that records metrics for gRPC requests.
func GRPCMetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	metrics.IncrCounter([]string{"grpc", "requests", info.FullMethod}, 1)
	metrics.MeasureSince([]string{"grpc", "requests", info.FullMethod, "latency"}, start)
	if err != nil {
		metrics.IncrCounter([]string{"grpc", "requests", info.FullMethod, "error"}, 1)
	}
	return resp, err
}
