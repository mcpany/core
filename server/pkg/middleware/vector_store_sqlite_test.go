// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSQLiteVectorStore(t *testing.T) {
	// Create temp db
	f, err := os.CreateTemp("", "semantic_cache_test_*.db")
	assert.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	store, err := NewSQLiteVectorStore(dbPath)
	assert.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	key := "test_tool"
	vec1 := []float32{1.0, 0.0, 0.0}
	res1 := "result1"

	// 1. Add
	err = store.Add(ctx, key, vec1, res1, 1*time.Minute)
	assert.NoError(t, err)

	// 2. Search exact
	res, score, found, err := store.Search(ctx, key, vec1)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.GreaterOrEqual(t, score, float32(0.99))
	assert.Equal(t, res1, res)

	// 3. Search similar
	vec2 := []float32{0.99, 0.05, 0.0}
	res, score, found, err = store.Search(ctx, key, vec2)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Greater(t, score, float32(0.9)) // High similarity
	assert.Equal(t, res1, res)

	// 4. Search orthogonal (should return match but low score if it's the only one? No, we iterate all and find BEST match)
	// Actually current implementation finds BEST match.
	// We might get result1 with score 0.
	vec3 := []float32{0.0, 1.0, 0.0}
	res, score, found, err = store.Search(ctx, key, vec3)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Less(t, score, float32(0.1))
	assert.Equal(t, res1, res)
}

func TestSQLiteVectorStore_Expiry(t *testing.T) {
	f, err := os.CreateTemp("", "semantic_cache_test_*.db")
	assert.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	store, err := NewSQLiteVectorStore(dbPath)
	assert.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	key := "test_tool"
	vec1 := []float32{1.0, 0.0, 0.0}
	res1 := "result1"

	err = store.Add(ctx, key, vec1, res1, 1*time.Millisecond)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	_, _, found, err := store.Search(ctx, key, vec1)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestSQLiteVectorStore_Persistence(t *testing.T) {
	f, err := os.CreateTemp("", "semantic_cache_test_*.db")
	assert.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Instance 1
	store1, err := NewSQLiteVectorStore(dbPath)
	assert.NoError(t, err)

	ctx := context.Background()
	key := "test_tool"
	vec1 := []float32{1.0, 0.0, 0.0}
	res1 := "result1"

	err = store1.Add(ctx, key, vec1, res1, 10*time.Minute)
	assert.NoError(t, err)
	store1.Close()

	// Instance 2 (Re-open)
	store2, err := NewSQLiteVectorStore(dbPath)
	assert.NoError(t, err)
	defer store2.Close()

	res, _, found, err := store2.Search(ctx, key, vec1)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, res1, res)
}
