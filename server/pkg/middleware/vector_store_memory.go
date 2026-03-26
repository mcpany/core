// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"math"
	"sync"
	"time"
)

// SimpleVectorStore is a thread-safe, in-memory storage for embedding vectors and their associated results.
//
// Summary: Implements a basic in-memory vector database using cosine similarity.
type SimpleVectorStore struct {
	mu         sync.RWMutex
	items      map[string][]*VectorEntry
	maxEntries int
}

// VectorEntry represents a single data point in the vector store, including its normalized vector and metadata.
//
// Summary: Encapsulates a stored vector and its cached result.
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

// NewSimpleVectorStore initializes and returns a new SimpleVectorStore instance.
//
// Summary: Factory function that initializes a thread-safe, in-memory vector store with a default entry limit to prevent excessive memory consumption.
//
// Returns:
//   - *SimpleVectorStore: A pointer to the initialized vector store.
//
// Side Effects:
//   - Allocates memory for the internal item map.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		items:      make(map[string][]*VectorEntry),
		maxEntries: 100, // Limit per key to prevent OOM
	}
}

// Add inserts or updates a vector entry for a specific key, performing normalization and expiration cleanup.
//
// Summary: Adds a vector and its associated result to the store.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - key (string): The logical category or identifier for the vector.
//   - vector ([]float32): The raw embedding vector.
//   - result (any): The data to be associated with this vector.
//   - ttl (time.Duration): Duration until the entry expires.
//
// Returns:
//   - error: Nil, as in-memory additions are generally infallible.
//
// Side Effects:
//   - Modifies the internal items map.
//   - May evict the oldest entry for the key if the maxEntries limit is reached.
//
// Errors:
//   - Returns an error if the operation fails.
func (s *SimpleVectorStore) Add(_ context.Context, key string, vector []float32, result any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanup(key)

	entries := s.items[key]
	if len(entries) >= s.maxEntries {
		// Evict oldest (FIFO)
		entries = entries[1:]
	}

	// ⚡ BOLT: Normalize vector before storage to avoid Sqrt/Div in Search loop
	// Randomized Selection from Top 5 High-Impact Targets
	normalizedVector, norm := normalize(vector)

	entry := &VectorEntry{
		Vector:    normalizedVector,
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
		Norm:      norm,
	}
	s.items[key] = append(entries, entry)
	return nil
}

// Search performs a cosine similarity search against all active entries for a given key.
//
// Summary: Calculates the dot product between the normalized query vector and all non-expired stored vectors for the given key, returning the result with the highest similarity.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - key (string): The key to search within.
//   - query ([]float32): The query vector to compare against.
//
// Returns:
//   - any: The result associated with the most similar vector.
//   - float32: The cosine similarity score of the best match.
//   - bool: True if a non-expired match was found, false otherwise.
//
// Side Effects:
//   - Performs multiple dot product calculations over the stored vectors.
//
// Errors:
//   - None.
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

	// Normalize query once
	normalizedQuery, _ := normalize(query)

	for _, entry := range entries {
		if now.After(entry.ExpiresAt) {
			continue
		}
		// Since both vectors are normalized, dot product == cosine similarity
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

// Prune manually triggers a cleanup of expired entries for a specific key.
//
// Summary: Removes stale vectors from the store.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - key (string): The key to clean up.
//
// Side Effects:
//   - Modifies the internal items map by removing expired entries.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
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

func normalize(v []float32) ([]float32, float32) {
	norm := vectorNorm(v)
	if norm == 0 {
		// Return copy to avoid side effects if caller modifies result
		// or if we decide to normalize in place later
		// For now, return original slice reference if unchanged?
		// Safest is to return a copy if we modify, but if we don't modify, original is fine.
		// Let's return the original reference if norm is 0 to avoid alloc.
		return v, 0
	}

	normalized := make([]float32, len(v))
	invNorm := 1.0 / norm
	for i, x := range v {
		normalized[i] = x * invNorm
	}
	return normalized, norm
}

func dotProduct(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sum float32
	// ⚡ BOLT: Simple dot product loop without bounds checking in loop if possible
	// Go compiler does eliminate bounds checks for `range`.
	for i, v := range a {
		sum += v * b[i]
	}
	return sum
}
