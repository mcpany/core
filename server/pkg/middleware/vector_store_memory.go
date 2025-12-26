// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"math"
	"sync"
	"time"
)

// SimpleVectorStore is a naive in-memory vector store.
// It stores vectors in memory and supports searching by cosine similarity.
// It also handles expiration of entries based on a TTL.
type SimpleVectorStore struct {
	mu         sync.RWMutex
	items      map[string][]*VectorEntry
	maxEntries int
}

// VectorEntry represents a single entry in the vector store.
// It contains the vector embedding, the cached result, and metadata for expiration and similarity calculation.
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
//   - A pointer to a new SimpleVectorStore instance.
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
//   - key: The identifier for the group of entries.
//   - vector: The embedding vector associated with the result.
//   - result: The data to be cached.
//   - ttl: The time-to-live duration for the entry.
//
// Returns:
//   - An error if the operation fails (currently always returns nil).
func (s *SimpleVectorStore) Add(key string, vector []float32, result any, ttl time.Duration) error {
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
// Parameters:
//   - key: The identifier for the group of entries to search.
//   - query: The query vector to compare against stored vectors.
//
// Returns:
//   - The cached result if a match is found, otherwise nil.
//   - The cosine similarity score (between -1 and 1).
//   - A boolean indicating if a valid match was found (true) or not (false).
func (s *SimpleVectorStore) Search(key string, query []float32) (any, float32, bool) {
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

// Prune removes expired entries for the specified key from the vector store.
//
// key is the identifier for the group of entries to prune.
func (s *SimpleVectorStore) Prune(key string) {
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
