// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InspectorMiddleware logs JSON-RPC traffic for the Inspector UI.
func InspectorMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		start := time.Now()

		// Execute the handler
		result, err := next(ctx, method, req)
		duration := time.Since(start)

		// Async logging to avoid blocking the critical path too much
		// (Though logging.GetLogger().Info is already somewhat async/buffered in our setup)
		go func() {
			logger := logging.GetLogger()

			// Prepare payload
			payload := map[string]any{
				"method":   method,
				"duration": duration.String(),
				"request":  req,
			}

			if err != nil {
				payload["error"] = err.Error()
			} else {
				payload["result"] = result
			}

			// We marshal the payload to JSON string so it can be passed as the "message" or "payload" attribute.
			// However, our BroadcastHandler puts attributes into Metadata.
			// So we can just log with attributes.

			// But wait, the LogEntry message is a string. If we put the whole JSON object in attributes,
			// the UI will receive it in `metadata`.

			// Let's Log:
			// Level: INFO
			// Source: INSPECTOR
			// Message: "Traffic <Method>"
			// Attributes: payload=<JSON_OBJECT>

			// Since slog attributes values are interface{}, we can pass the map directly.
			// The JSONHandler/BroadcastHandler will marshal it.

			logger.Info(
				"RPC Traffic",
				"source", "INSPECTOR",
				"method", method,
				"payload", payload,
			)
		}()

		return result, err
	}
}
