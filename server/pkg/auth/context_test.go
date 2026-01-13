// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextHelpers(t *testing.T) {
	t.Run("user_context", func(t *testing.T) {
		ctx := context.Background()
		userID := "test-user"

		// Test ContextWithUser
		ctx = ContextWithUser(ctx, userID)

		// Test UserFromContext
		got, ok := UserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, userID, got)

		// Test missing user
		_, ok = UserFromContext(context.Background())
		assert.False(t, ok)
	})

	t.Run("profile_context", func(t *testing.T) {
		ctx := context.Background()
		profileID := "test-profile"

		// Test ContextWithProfileID
		ctx = ContextWithProfileID(ctx, profileID)

		// Test ProfileIDFromContext
		got, ok := ProfileIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, profileID, got)

		// Test missing profile
		_, ok = ProfileIDFromContext(context.Background())
		assert.False(t, ok)
	})

	t.Run("api_key", func(t *testing.T) {
		ctx := context.Background()
		apiKey := "test-api-key"

		// Test ContextWithAPIKey
		ctx = ContextWithAPIKey(ctx, apiKey)

		// Test APIKeyFromContext
		got, ok := APIKeyFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, apiKey, got)

		// Test missing api key
		_, ok = APIKeyFromContext(context.Background())
		assert.False(t, ok)
	})
}
