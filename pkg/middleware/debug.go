// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"

	"github.com/mcpany/core/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DebugMiddleware returns a middleware function that logs the full request and
// response of each MCP method call. This is useful for debugging and
// understanding the flow of data through the server.
func DebugMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			log := logging.GetLogger()

			reqBytes, err := json.Marshal(req)
			if err != nil {
				log.Error("Failed to marshal request for debugging", "error", err)
			} else {
				log.Debug("MCP Request", "method", method, "request", string(reqBytes))
			}

			result, err := next(ctx, method, req)
			if err != nil {
				log.Error("MCP method failed", "method", method, "error", err)
				return nil, err
			}

			resBytes, err := json.Marshal(result)
			if err != nil {
				log.Error("Failed to marshal response for debugging", "error", err)
			} else {
				log.Debug("MCP Response", "method", method, "response", string(resBytes))
			}

			return result, nil
		}
	}
}
