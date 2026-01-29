package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleVectorStore_Add_Eviction(t *testing.T) {
	store := NewSimpleVectorStore()
	store.maxEntries = 2 // Set small limit for testing

	ctx := context.Background()
	key := "test_key"
	vec := []float32{1.0, 0.0, 0.0}

	// Add 3 entries
	err := store.Add(ctx, key, vec, "item1", time.Hour)
	assert.NoError(t, err)
	err = store.Add(ctx, key, vec, "item2", time.Hour)
	assert.NoError(t, err)
	err = store.Add(ctx, key, vec, "item3", time.Hour)
	assert.NoError(t, err)

	store.mu.RLock()
	entries := store.items[key]
	store.mu.RUnlock()

	// Should have 2 entries
	assert.Equal(t, 2, len(entries))
	// FIFO eviction: item1 should be gone, item2 and item3 remain
	assert.Equal(t, "item2", entries[0].Result)
	assert.Equal(t, "item3", entries[1].Result)
}

func TestSimpleVectorStore_Search(t *testing.T) {
	store := NewSimpleVectorStore()
	ctx := context.Background()
	key := "test_key"

	// Add entries
	// vec1: [1, 0]
	// vec2: [0, 1]
	err := store.Add(ctx, key, []float32{1.0, 0.0}, "vec1", time.Hour)
	assert.NoError(t, err)
	err = store.Add(ctx, key, []float32{0.0, 1.0}, "vec2", time.Hour)
	assert.NoError(t, err)

	// Search close to vec1
	query := []float32{0.9, 0.1}
	res, score, found := store.Search(ctx, key, query)
	assert.True(t, found)
	assert.Equal(t, "vec1", res)
	assert.Greater(t, score, float32(0.0))

	// Search close to vec2
	query2 := []float32{0.1, 0.9}
	res, score, found = store.Search(ctx, key, query2)
	assert.True(t, found)
	assert.Equal(t, "vec2", res)
	assert.Greater(t, score, float32(0.0))

	// Search non-existent key
	_, _, found = store.Search(ctx, "other_key", query)
	assert.False(t, found)
}

func TestSimpleVectorStore_Expiry(t *testing.T) {
	store := NewSimpleVectorStore()
	ctx := context.Background()
	key := "test_key"

	// Add expired entry
	err := store.Add(ctx, key, []float32{1.0}, "expired", -time.Hour)
	assert.NoError(t, err)

	// Search should ignore it
	_, _, found := store.Search(ctx, key, []float32{1.0})
	assert.False(t, found)
}

func TestSimpleVectorStore_Prune(t *testing.T) {
	store := NewSimpleVectorStore()
	ctx := context.Background()
	key := "test_key"

	// Add one valid, one expired
	err := store.Add(ctx, key, []float32{1.0}, "valid", time.Hour)
	assert.NoError(t, err)
	err = store.Add(ctx, key, []float32{1.0}, "expired", -time.Hour)
	assert.NoError(t, err)

	store.Prune(ctx, key)

	store.mu.RLock()
	entries := store.items[key]
	store.mu.RUnlock()

	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "valid", entries[0].Result)
}

func TestVectorMath(t *testing.T) {
	// vectorNorm
	v := []float32{3.0, 4.0}
	assert.Equal(t, float32(5.0), vectorNorm(v))

	// cosineSimilarityOptimized
	v1 := []float32{1.0, 0.0}
	v2 := []float32{1.0, 0.0}
	norm1 := vectorNorm(v1)
	norm2 := vectorNorm(v2)
	assert.Equal(t, float32(1.0), cosineSimilarityOptimized(v1, v2, norm1, norm2))

	v3 := []float32{0.0, 1.0}
	norm3 := vectorNorm(v3)
	assert.Equal(t, float32(0.0), cosineSimilarityOptimized(v1, v3, norm1, norm3))

	// Edge cases
	assert.Equal(t, float32(0.0), cosineSimilarityOptimized([]float32{}, []float32{}, 0, 0))
	assert.Equal(t, float32(0.0), cosineSimilarityOptimized(v1, []float32{1.0}, norm1, 1.0)) // Length mismatch
	assert.Equal(t, float32(0.0), cosineSimilarityOptimized(v1, v2, 0, norm2)) // Zero norm
}

func TestSimpleVectorStore_Concurrency(t *testing.T) {
	store := NewSimpleVectorStore()
	ctx := context.Background()
	key := "concurrent_key"

	// Concurrent Adds
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(i int) {
			store.Add(ctx, key, []float32{float32(i)}, fmt.Sprintf("item%d", i), time.Hour)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	store.mu.RLock()
	assert.LessOrEqual(t, len(store.items[key]), store.maxEntries)
	store.mu.RUnlock()
}
