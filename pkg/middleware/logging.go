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
	"log/slog"
	"time"

	"github.com/mcpany/core/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LoggingMiddleware creates an MCP middleware that logs information about each
// incoming request. It records the start and completion of each request,
// including the duration of the handling.
//
// This is useful for debugging and monitoring the flow of requests through the
// server.
//
// Parameters:
//   - log: The logger to be used. If `nil`, the default global logger will be
//     used.
//
// Returns an `mcp.Middleware` function.
func LoggingMiddleware(log *slog.Logger) mcp.Middleware {
	if log == nil {
		log = logging.GetLogger()
	}
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			log.Info("Request received", "method", method)
			result, err := next(ctx, method, req)
			log.Info("Request completed", "method", method, "duration", time.Since(start))
			return result, err
		}
	}
}
