package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockEmbeddingProvider struct {
	embeddings map[string][]float32
}

func (m *MockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if val, ok := m.embeddings[text]; ok {
		return val, nil
	}
	// Return a default zero vector or similar for misses? Or error?
	// For test we assume strict mapping
	return []float32{0, 0, 0}, nil
}

func TestSemanticCache(t *testing.T) {
	provider := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"hello":   {1.0, 0.0, 0.0},
			"hi":      {0.99, 0.05, 0.0}, // Very similar
			"goodbye": {0.0, 1.0, 0.0},   // Orthogonal
		},
	}

	cache := NewSemanticCache(provider, nil, 0.9) // Threshold 0.9

	ctx := context.Background()
	key := "test_tool"

	// 1. Set "hello" -> "world"
	emb, err := provider.Embed(ctx, "hello")
	assert.NoError(t, err)
	err = cache.Set(ctx, key, emb, "world", 1*time.Minute)
	assert.NoError(t, err)

	// 2. Get "hello" (Exact match)
	val, _, hit, err := cache.Get(ctx, key, "hello")
	assert.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "world", val)

	// 3. Get "hi" (Semantic match)
	val, _, hit, err = cache.Get(ctx, key, "hi")
	assert.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "world", val)

	// 4. Get "goodbye" (Miss)
	val, _, hit, err = cache.Get(ctx, key, "goodbye")
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, val)
}

func TestSemanticCache_Expiry(t *testing.T) {
	provider := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"hello": {1.0, 0.0, 0.0},
		},
	}

	cache := NewSemanticCache(provider, nil, 0.9)
	ctx := context.Background()
	key := "test_tool"

	// Set with short expiry
	emb, err := provider.Embed(ctx, "hello")
	assert.NoError(t, err)
	cache.Set(ctx, key, emb, "world", 1*time.Millisecond)

	time.Sleep(10 * time.Millisecond)

	// Get should fail
	val, _, hit, err := cache.Get(ctx, key, "hello")
	assert.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, val)
}
