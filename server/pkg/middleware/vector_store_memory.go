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
//
// Summary: Initializes an in-memory vector store with default limits.
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
//
// Summary: Stores a vector and result in memory, enforcing FIFO eviction if the per-key limit is reached.
//
// Parameters:
//   - ctx: context.Context. The context (unused).
//   - key: string. The partition key.
//   - vector: []float32. The embedding vector.
//   - result: any. The result to cache.
//   - ttl: time.Duration. The time-to-live for the entry.
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

	entry := &VectorEntry{
		Vector:    vector,
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
		Norm:      vectorNorm(vector),
	}
	s.items[key] = append(entries, entry)
	return nil
}

// Search searches for the most similar entry in the vector store for the given key and query vector.
//
// Summary: Performs a brute-force cosine similarity search across vectors associated with the key.
//
// Parameters:
//   - ctx: context.Context. The context (unused).
//   - key: string. The partition key.
//   - query: []float32. The query embedding vector.
//
// Returns:
//   - any: The best matching cached result.
//   - float32: The similarity score (cosine similarity).
//   - bool: True if a match was found.
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
	queryNorm := vectorNorm(query)

	for _, entry := range entries {
		if now.After(entry.ExpiresAt) {
			continue
		}
		score := cosineSimilarityOptimized(query, entry.Vector, queryNorm, entry.Norm)
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
// Summary: Removes entries that have exceeded their expiration time.
//
// Parameters:
//   - ctx: context.Context. The context (unused).
//   - key: string. The partition key.
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

func cosineSimilarityOptimized(a, b []float32, normA, normB float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	if normA == 0 || normB == 0 {
		return 0
	}

	var dotProduct float32
	for i := range a {
		dotProduct += a[i] * b[i]
	}
	return dotProduct / (normA * normB)
}
