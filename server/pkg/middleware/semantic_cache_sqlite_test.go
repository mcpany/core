// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteVectorStore(t *testing.T) {
	// Setup temporary DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectors.db")

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	key := "tool_a"
	vec := []float32{1.0, 0.0}
	result := map[string]interface{}{"foo": "bar"}
	ttl := 1 * time.Minute

	// Add entry
	err = store.Add(context.Background(), key, vec, result, ttl)
	assert.NoError(t, err)

	// Search in same session (should be in memory)
	res, score, found := store.Search(context.Background(), key, []float32{1.0, 0.0})
	assert.True(t, found)
	assert.Equal(t, float32(1.0), score)
	assert.Equal(t, result, res)

	// Close and reopen to test persistence
	err = store.Close()
	assert.NoError(t, err)

	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Search in new session (should be loaded from DB)
	res, score, found = store2.Search(context.Background(), key, []float32{1.0, 0.0})
	assert.True(t, found)
	assert.Equal(t, float32(1.0), score)

	// Check result content - we need to handle map[string]interface{} equality carefully if JSON marshaling changes types (e.g. numbers)
	// result is {"foo": "bar"}, which should be stable.
	assert.Equal(t, result, res)
}

func TestSQLiteVectorStore_Expiry(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectors_expiry.db")

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	key := "tool_b"
	vec := []float32{0.0, 1.0}
	result := "expired"
	ttl := 1 * time.Millisecond

	err = store.Add(context.Background(), key, vec, result, ttl)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Search should miss in memory (memory store handles expiry check)
	_, _, found := store.Search(context.Background(), key, vec)
	assert.False(t, found)

	// Close and reopen
	store.Close()

	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Should not load expired item
	_, _, found = store2.Search(context.Background(), key, vec)
	assert.False(t, found)
}
