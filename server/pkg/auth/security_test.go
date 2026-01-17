package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalAPIKey_QueryParam_Leak(t *testing.T) {
	// Setup
	manager := auth.NewManager()
	apiKey := "secret-key-123"
	manager.SetAPIKey(apiKey)

	// Test Case 1: Authenticate via Header (Should always work)
	t.Run("Authenticate via Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?foo=bar", nil)
		req.Header.Set("X-API-Key", apiKey)

		ctx, err := manager.Authenticate(context.Background(), "some-service", req)
		require.NoError(t, err)

		val, ok := auth.APIKeyFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, apiKey, val)
	})

	// Test Case 2: Authenticate via Query Param (Should FAIL now)
	t.Run("Authenticate via Query Param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?api_key="+apiKey, nil)

		_, err := manager.Authenticate(context.Background(), "some-service", req)

		// NOW: Expect Error
		assert.Error(t, err, "Should fail when using query param for global authentication")
		assert.Contains(t, err.Error(), "unauthorized")
	})
}
