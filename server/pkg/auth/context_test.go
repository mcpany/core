// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	t.Run("APIKey", func(t *testing.T) {
		apiKey := "test-api-key"
		ctxWithKey := ContextWithAPIKey(ctx, apiKey)
		val, ok := APIKeyFromContext(ctxWithKey)
		assert.True(t, ok)
		assert.Equal(t, apiKey, val)

		_, ok = APIKeyFromContext(ctx)
		assert.False(t, ok)
	})

	t.Run("User", func(t *testing.T) {
		userID := "test-user-id"
		ctxWithUser := ContextWithUser(ctx, userID)
		val, ok := UserFromContext(ctxWithUser)
		assert.True(t, ok)
		assert.Equal(t, userID, val)

		_, ok = UserFromContext(ctx)
		assert.False(t, ok)
	})

	t.Run("ProfileID", func(t *testing.T) {
		profileID := "test-profile-id"
		ctxWithProfile := ContextWithProfileID(ctx, profileID)
		val, ok := ProfileIDFromContext(ctxWithProfile)
		assert.True(t, ok)
		assert.Equal(t, profileID, val)

		_, ok = ProfileIDFromContext(ctx)
		assert.False(t, ok)
	})
}
