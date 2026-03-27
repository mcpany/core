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
// Summary: Represents a SimpleVectorStore.
type SimpleVectorStore struct {
	mu         sync.RWMutex
	items      map[string][]*VectorEntry
	maxEntries int
}

// VectorEntry represents a single entry in the vector store.
//
// Summary: Represents a VectorEntry.
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

// NewSimpleVectorStore provides newsimplevectorstore functionality.
//
// Summary: NewSimpleVectorStore.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		items:      make(map[string][]*VectorEntry),
		maxEntries: 100, // Limit per key to prevent OOM
	}
}

// Add provides add functionality.
//
// Summary: Add.
//
// Parameters.
//   - _: The parameter.
//   - key: The parameter.
//   - vector: The parameter.
//   - result: The parameter.
//   - ttl: The parameter.
//
// Returns.
//   - result: The result.
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

// Search provides search functionality.
//
// Summary: Search.
//
// Parameters.
//   - _: The parameter.
//   - key: The parameter.
//   - query: The parameter.
//   - float32: The parameter.
//   - bool: The parameter.
//
// Returns.
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

// Prune provides prune functionality.
//
// Summary: Prune.
//
// Parameters.
//   - _: The parameter.
//   - key: The parameter.
//
// Returns.
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
