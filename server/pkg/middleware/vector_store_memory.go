// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"math"
	"sync"
	"time"
)

// SimpleVectorStore is a naive in-memory vector store.
type SimpleVectorStore struct {
	mu         sync.RWMutex
	items      map[string][]*VectorEntry
	maxEntries int
}

// VectorEntry represents a single entry in the vector store.
type VectorEntry struct {
	// Vector is the embedding vector.
	Vector []float32
	// Result is the cached result associated with the vector.
	Result any
	// ExpiresAt is the timestamp when this entry expires.
	ExpiresAt time.Time
	// Norm is the precomputed Euclidean norm of the vector.
	Norm float32
}

// NewSimpleVectorStore creates a new SimpleVectorStore.
// It initializes the store with a default configuration.
//
// Returns:
//   - *SimpleVectorStore: A pointer to the newly created SimpleVectorStore.
func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		items:      make(map[string][]*VectorEntry),
		maxEntries: 100, // Limit per key to prevent OOM
	}
}

// Add adds a new entry to the vector store.
// It evicts the oldest entry if the store exceeds the maximum number of entries for the key.
//
// Parameters:
//   - key: The key associated with the entry.
//   - vector: The embedding vector.
//   - result: The result to cache.
//   - ttl: The time-to-live for the entry.
//
// Returns:
//   - error: An error if the operation fails (currently always nil).
func (s *SimpleVectorStore) Add(_ context.Context, key string, vector []float32, result any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanup(key)

	entries := s.items[key]
	if len(entries) >= s.maxEntries {
		// Evict oldest (FIFO)
		entries = entries[1:]
	}

	// ⚡ BOLT: Normalize vector on insertion to optimize search speed.
	// Randomized Selection from Top 5 High-Impact Targets
	normalizedVector := normalize(vector)

	entry := &VectorEntry{
		Vector:    normalizedVector,
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
		Norm:      1.0, // Pre-normalized
	}
	s.items[key] = append(entries, entry)
	return nil
}

// Search searches for the most similar entry in the vector store for the given key and query vector.
// It returns the result, the similarity score, and a boolean indicating if a match was found.
//
// Parameters:
//   - key: The key to search for.
//   - query: The query vector.
//
// Returns:
//   - any: The cached result if found.
//   - float32: The similarity score (cosine similarity).
//   - bool: True if a match was found, false otherwise.
func (s *SimpleVectorStore) Search(_ context.Context, key string, query []float32) (any, float32, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.items[key]
	if !ok {
		return nil, 0, false
	}

	now := time.Now()
	var bestResult any
	var bestScore float32 = -1.0
	// ⚡ BOLT: Normalize query vector once to enable dot product optimization
	// Randomized Selection from Top 5 High-Impact Targets
	normalizedQuery := normalize(query)

	for _, entry := range entries {
		if now.After(entry.ExpiresAt) {
			continue
		}
		// entry.Vector is already normalized (Norm=1.0).
		// We use simple dot product.
		score := dotProduct(normalizedQuery, entry.Vector)
		if score > bestScore {
			bestScore = score
			bestResult = entry.Result
		}
	}

	if bestScore == -1.0 {
		return nil, 0, false
	}

	return bestResult, bestScore, true
}

// Prune removes expired entries from the vector store for the given key.
//
// Parameters:
//   - key: The key to prune entries for.
func (s *SimpleVectorStore) Prune(_ context.Context, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanup(key)
}

func (s *SimpleVectorStore) cleanup(key string) {
	entries, ok := s.items[key]
	if !ok {
		return
	}
	now := time.Now()
	// Filter in place
	n := 0
	for _, e := range entries {
		if now.Before(e.ExpiresAt) {
			entries[n] = e
			n++
		}
	}
	// Zero out the rest to help GC
	for i := n; i < len(entries); i++ {
		entries[i] = nil
	}
	s.items[key] = entries[:n]
}

func vectorNorm(v []float32) float32 {
	var sum float32
	for _, x := range v {
		sum += x * x
	}
	return float32(math.Sqrt(float64(sum)))
}

func normalize(v []float32) []float32 {
	norm := vectorNorm(v)
	if norm == 0 {
		// Return copy of zero vector to avoid mutating original if we were mutating in place (we aren't, but safety)
		// Or just return v if we treat it as immutable
		return append([]float32(nil), v...)
	}
	normalized := make([]float32, len(v))
	for i, x := range v {
		normalized[i] = x / norm
	}
	return normalized
}

func dotProduct(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float32
	for i := range a {
		dotProduct += a[i] * b[i]
	}
	return dotProduct
}
