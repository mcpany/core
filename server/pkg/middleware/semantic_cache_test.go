// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockProvider for testing
type MockProvider struct {
	embeddings map[string][]float32
}

func (m *MockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if emb, ok := m.embeddings[text]; ok {
		return emb, nil
	}
	return []float32{0, 0, 0}, nil
}

func TestSemanticCache(t *testing.T) {
	provider := &MockProvider{
		embeddings: map[string][]float32{
			"hello world":  {1.0, 0.0, 0.0},
			"hello world!": {0.99, 0.01, 0.0}, // Very similar
			"goodbye":      {0.0, 1.0, 0.0},   // Orthogonal
		},
	}

	cache := NewSemanticCache(provider, 0.9)
	ctx := context.Background()
	key := "test-tool"

	// 1. Set cache
	err := cache.Set(ctx, key, "hello world", "result_hello", time.Minute)
	assert.NoError(t, err)

	// 2. Get exact match
	res, hit, err := cache.Get(ctx, key, "hello world")
	assert.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "result_hello", res)

	// 3. Get similar match
	res, hit, err = cache.Get(ctx, key, "hello world!")
	assert.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "result_hello", res)

	// 4. Get dissimilar match
	res, hit, err = cache.Get(ctx, key, "goodbye")
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, res)
}

func TestSimpleVectorStore_Expiry(t *testing.T) {
	store := NewSimpleVectorStore()
	key := "tool"
	vec := []float32{1, 0}

	// Add item with short TTL
	store.Add(key, vec, "result", 100*time.Millisecond)

	// Search immediately
	res, _, found := store.Search(key, vec)
	assert.True(t, found)
	assert.Equal(t, "result", res)

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Search again
	res, _, found = store.Search(key, vec)
	assert.False(t, found)
	assert.Nil(t, res)
}
