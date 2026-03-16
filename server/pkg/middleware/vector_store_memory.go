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
//
// Summary: SimpleVectorStore is a naive in-memory vector store.
//
// Summary: SimpleVectorStore is a naive in-memory vector store.
type SimpleVectorStore struct {
	mu         sync.RWMutex
	items      map[string][]*VectorEntry
	maxEntries int
// VectorEntry represents a single entry in the vector store.
//
// Summary: VectorEntry represents a single entry in the vector store.

// VectorEntry represents a single entry in the vector store.
//
// Summary: VectorEntry represents a single entry in the vector store.
type VectorEntry struct {
	// Vector is the embedding vector.
	Vector []float32
	// Result is the cached result associated with the vector.
	Result any
	// ExpiresAt is the timestamp when this entry expires.
	ExpiresAt time.Time
// NewSimpleVectorStore creates a new SimpleVectorStore.
//
// Summary: NewSimpleVectorStore creates a new SimpleVectorStore.
//
// Parameters:
//   - None.
//
// Returns:
//   - *SimpleVectorStore: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Summary: NewSimpleVectorStore creates a new SimpleVectorStore.
//
// Parameters:
//   - None.
//
// Returns:
// Add adds a new entry to the vector store.
//
// Summary: Add adds a new entry to the vector store.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - key (string): The textual representation of key.
//   - vector ([]float32): The numeric value for vector.
//   - result (any): The provided result data.
//   - ttl (time.Duration): The provided ttl data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	}
}

// Add adds a new entry to the vector store.
//
// Summary: Add adds a new entry to the vector store.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - key (string): The textual representation of key.
//   - vector ([]float32): The numeric value for vector.
//   - result (any): The provided result data.
//   - ttl (time.Duration): The provided ttl data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *SimpleVectorStore) Add(_ context.Context, key string, vector []float32, result any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

// Search searches for the most similar entry in the vector store for the given key and query vector.
//
// Summary: Search searches for the most similar entry in the vector store for the given key and query vector.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - key (string): The textual representation of key.
//   - query ([]float32): The numeric value for query.
//
// Returns:
//   - any: The resulting object or data structure.
//   - float32: The calculated numeric value.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.

	entry := &VectorEntry{
		Vector:    normalizedVector,
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
		Norm:      norm,
	}
	s.items[key] = append(entries, entry)
	return nil
}

// Search searches for the most similar entry in the vector store for the given key and query vector.
//
// Summary: Search searches for the most similar entry in the vector store for the given key and query vector.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - key (string): The textual representation of key.
//   - query ([]float32): The numeric value for query.
//
// Returns:
//   - any: The resulting object or data structure.
//   - float32: The calculated numeric value.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *SimpleVectorStore) Search(_ context.Context, key string, query []float32) (any, float32, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.items[key]
// Prune removes expired entries from the vector store for the given key.
//
// Summary: Prune removes expired entries from the vector store for the given key.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - key (string): The textual representation of key.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// Prune removes expired entries from the vector store for the given key.
//
// Summary: Prune removes expired entries from the vector store for the given key.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - key (string): The textual representation of key.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
