// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetSecret(t *testing.T) {
	mock := NewMockProvider()
	mock.Secrets["my-secret"] = "top-secret-value"

	manager := NewManager(mock)
	ctx := context.Background()

	// 1. Initial fetch (cache miss)
	val, err := manager.GetSecret(ctx, "my-secret")
	assert.NoError(t, err)
	assert.Equal(t, "top-secret-value", val)

	// 2. Fetch from cache (mock update shouldn't be reflected yet if cached)
	mock.Secrets["my-secret"] = "changed-in-backend"
	val, err = manager.GetSecret(ctx, "my-secret")
	assert.NoError(t, err)
	assert.Equal(t, "top-secret-value", val)

	// 3. Not found
	_, err = manager.GetSecret(ctx, "unknown-secret")
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestManager_RotateSecret(t *testing.T) {
	mock := NewMockProvider()
	mock.Secrets["db-password"] = "old-password"

	manager := NewManager(mock)
	ctx := context.Background()

	// Rotate
	newVal, err := manager.RotateSecret(ctx, "db-password")
	assert.NoError(t, err)
	assert.Equal(t, "old-password_rotated", newVal)

	// Verify cache is updated
	val, err := manager.GetSecret(ctx, "db-password")
	assert.NoError(t, err)
	assert.Equal(t, "old-password_rotated", val)
}
