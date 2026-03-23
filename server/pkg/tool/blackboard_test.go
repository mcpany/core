// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlackboardTool(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "blackboard.db")

	b, err := NewBlackboardTool(dbPath)
	require.NoError(t, err)
	defer b.Close()

	ctx := context.Background()

	// Test Get non-existent
	val, err := b.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, "", val)

	// Test Set
	err = b.Set(ctx, "testkey", "testvalue")
	assert.NoError(t, err)

	// Test Get existing
	val, err = b.Get(ctx, "testkey")
	assert.NoError(t, err)
	assert.Equal(t, "testvalue", val)

	// Test Update
	err = b.Set(ctx, "testkey", "newvalue")
	assert.NoError(t, err)

	val, err = b.Get(ctx, "testkey")
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", val)

	// Ensure file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}
