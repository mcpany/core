// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_PromptsGet_ServiceAuth(t *testing.T) {
	// Setup Auth Manager with:
	// 1. Global API Key: "global-secret"
	// 2. Service "secure" with API Key: "service-secret"
	authManager := auth.NewManager()
	authManager.SetAPIKey("global-secret")

	authenticator := &auth.APIKeyAuthenticator{
		ParamName: "X-API-Key",
		Value:     "service-secret",
		In:        configv1.APIKeyAuth_HEADER,
	}
	// "secure" service requires "service-secret"
	err := authManager.AddAuthenticator("secure", authenticator)
	require.NoError(t, err)

	mw := middleware.AuthMiddleware(authManager)

	// Mock next handler
	var nextCalled bool
	nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.GetPromptResult{}, nil
	}

	handler := mw(nextHandler)

	t.Run("should fail when accessing secure prompt with only global key", func(t *testing.T) {
		nextCalled = false
		// Attempt to access "secure.prompt"
		req := &mcp.GetPromptRequest{
			Params: &mcp.GetPromptParams{
				Name: "secure.prompt",
			},
		}

		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		// We provide ONLY the Global Key.
		// If service auth is enforced, this should fail (because service "secure" expects "service-secret").
		// If service auth is bypassed, this will succeed (because Global Key is valid).
		httpReq.Header.Set("X-API-Key", "global-secret")

		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// This call should fail if security is working correctly
		_, err = handler(ctx, consts.MethodPromptsGet, req)

		// Assert that we get an error (Unauthorized).
		// Currently this fails (err is nil).
		if err == nil {
			t.Logf("VULNERABILITY DETECTED: Service auth bypassed using global key for prompt access. Err is nil.")
			assert.Error(t, err, "Should require service authentication, but bypassed with global key")
			assert.True(t, nextCalled, "Next handler should have been called (bypass)")
		} else {
			assert.Contains(t, err.Error(), "unauthorized")
			assert.False(t, nextCalled, "Next handler should not have been called")
		}
	})
}
