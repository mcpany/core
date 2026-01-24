// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusError   = "error"
	statusSuccess = "success"
)

// InspectorMiddleware intercepts MCP requests and logs them for the Inspector UI.
func InspectorMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		start := time.Now()

		// Execute the handler
		res, err := next(ctx, method, req)
		duration := time.Since(start)

		// Prepare log payload
		payload := map[string]any{
			"method":    method,
			"timestamp": start.Format(time.RFC3339),
			"duration":  duration.String(),
			"request":   req,
		}

		if err != nil {
			payload["error"] = err.Error()
			payload["status"] = statusError
		} else {
			payload["result"] = res
			payload["status"] = statusSuccess
		}

		// Log with source=INSPECTOR
		// We marshal the payload to a JSON string so the UI can parse it as a structured object
		// and display it nicely.
		if jsonBytes, err := json.Marshal(payload); err == nil {
			// Redact sensitive information before logging
			redactedBytes := util.RedactJSON(jsonBytes)
			logging.GetLogger().Info(string(redactedBytes), "source", "INSPECTOR")
		} else {
			logging.GetLogger().Error("Failed to marshal inspector payload", "error", err, "source", "INSPECTOR")
		}

		return res, err
	}
}
