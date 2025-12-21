// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"sync"
	"time"

	"github.com/mcpany/core/pkg/ai/embeddings"
)

// VectorEntry represents an item in the vector store.
type VectorEntry struct {
	ToolName  string
	Embedding embeddings.Embedding
	Value     any
	ExpiresAt time.Time
}

// VectorStore implements a simple in-memory vector store.
type VectorStore struct {
	entries []VectorEntry
	mu      sync.RWMutex
}

// NewVectorStore creates a new VectorStore.
func NewVectorStore() *VectorStore {
	return &VectorStore{
		entries: make([]VectorEntry, 0),
	}
}

// Add adds an entry to the vector store.
func (s *VectorStore) Add(toolName string, embedding embeddings.Embedding, value any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Amortized cleanup: only compact if size exceeds threshold
	if len(s.entries) > 1000 {
		now := time.Now()
		newEntries := make([]VectorEntry, 0, len(s.entries)/2)
		for _, e := range s.entries {
			if e.ExpiresAt.After(now) {
				newEntries = append(newEntries, e)
			}
		}
		s.entries = newEntries
	}

	s.entries = append(s.entries, VectorEntry{
		ToolName:  toolName,
		Embedding: embedding,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	})
}

// Search searches for the nearest neighbor embedding.
func (s *VectorStore) Search(toolName string, query embeddings.Embedding, threshold float32) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestMatch any
	var bestScore float32 = -1.0
	now := time.Now()

	for _, entry := range s.entries {
		if entry.ToolName != toolName {
			continue
		}
		if !entry.ExpiresAt.After(now) {
			continue
		}

		score := cosineSimilarity(query, entry.Embedding)
		if score > bestScore {
			bestScore = score
			bestMatch = entry.Value
		}
	}

	if bestScore >= threshold {
		return bestMatch, true
	}
	return nil, false
}

func cosineSimilarity(a, b embeddings.Embedding) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dotProduct float32
	for i := range a {
		dotProduct += a[i] * b[i]
	}
	// Vectors are assumed to be normalized
	return dotProduct
}

// Clear clears the vector store.
func (s *VectorStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make([]VectorEntry, 0)
}
