// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// TestAuthMiddleware_Bypass_WithGlobalKey tests that if a global API key is set,
// it is enforced even if no authenticator is registered for a specific service.
func TestAuthMiddleware_Bypass_WithGlobalKey(t *testing.T) {
	// 1. Setup Auth Manager with a global API key
	am := auth.NewManager()
	am.SetAPIKey("secret-global-key")

	// 2. Setup the middleware
	mw := AuthMiddleware(am)

	// 3. Create a dummy next handler that signifies success (access granted)
	nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Access Granted",
			},
		}}, nil
	}

	// 4. Wrap the next handler with the middleware
	handler := mw(nextHandler)

	// 5. Create a request for a service that DOES NOT have a specific authenticator
	// "unknown-service" has no authenticator registered in 'am'.
	ctx := context.Background()
	req, _ := http.NewRequest("POST", "http://127.0.0.1", nil)
	// Important: We DO NOT send the API Key header to simulate an unauthorized request
	// req.Header.Set("X-API-Key", "secret-global-key")

	// Add http.Request to context as required by middleware
	ctx = context.WithValue(ctx, HTTPRequestContextKey, req)

	// 6. Execute the handler
	// We expect this to FAIL with "unauthorized" because we didn't provide the global API key.
	// We pass nil as request params because our dummy handler doesn't use them
	_, err := handler(ctx, "unknown-service.some-method", nil)

	// 7. Verify the result
	// The correct secure behavior is that err should be non-nil ("unauthorized").
	assert.Error(t, err, "Expected unauthorized error due to missing global API key")
	if err != nil {
		assert.Contains(t, err.Error(), "unauthorized")
	}
}
