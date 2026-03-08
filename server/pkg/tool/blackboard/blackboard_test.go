// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package blackboard

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlackboardStore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "blackboard_test.db")

	store, err := NewBlackboardStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	err = store.Set(ctx, "agent1", "memory_key", "memory_value")
	assert.NoError(t, err)

	val, found, err := store.Get(ctx, "agent1", "memory_key")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "memory_value", val)

	err = store.Set(ctx, "agent2", "shared_state", map[string]interface{}{
		"foo": "bar",
		"num": float64(42),
	})
	assert.NoError(t, err)

	val, found, err = store.Get(ctx, "agent2", "shared_state")
	assert.NoError(t, err)
	assert.True(t, found)

	valMap, ok := val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "bar", valMap["foo"])
	assert.Equal(t, float64(42), valMap["num"])

	err = store.Delete(ctx, "agent1", "memory_key")
	assert.NoError(t, err)

	val, found, err = store.Get(ctx, "agent1", "memory_key")
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, val)
}
