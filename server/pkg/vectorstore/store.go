// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vectorstore

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Entry represents a stored vector and its associated data.
type Entry struct {
	Vector         []float32
	Data           any
	ExpirationTime time.Time
}

// Result represents a search result.
type Result struct {
	Data       any
	Similarity float32
}

// Store defines the interface for a vector store.
type Store interface {
	Add(vector []float32, data any, ttl time.Duration) error
	Search(vector []float32, limit int, threshold float32) ([]Result, error)
	Clear()
}

// SimpleStore is an in-memory implementation of Store.
type SimpleStore struct {
	entries     []Entry
	mu          sync.RWMutex
	maxCapacity int
}

// NewSimpleStore creates a new SimpleStore.
func NewSimpleStore(maxCapacity int) *SimpleStore {
	if maxCapacity <= 0 {
		maxCapacity = 1000 // Default safe limit
	}
	return &SimpleStore{
		entries:     make([]Entry, 0, maxCapacity),
		maxCapacity: maxCapacity,
	}
}

// Add adds a vector and data to the store.
func (s *SimpleStore) Add(vector []float32, data any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If full, try to remove expired first
	if len(s.entries) >= s.maxCapacity {
		now := time.Now()
		// Filter in place
		n := 0
		for _, e := range s.entries {
			if e.ExpirationTime.After(now) {
				s.entries[n] = e
				n++
			}
		}
		s.entries = s.entries[:n]

		// If still full, remove oldest (first element, FIFO assumption)
		if len(s.entries) >= s.maxCapacity {
			// Copy elements to shift left
			copy(s.entries, s.entries[1:])
			s.entries = s.entries[:len(s.entries)-1]
		}
	}

	expiration := time.Now().Add(ttl)
	s.entries = append(s.entries, Entry{
		Vector:         vector,
		Data:           data,
		ExpirationTime: expiration,
	})
	return nil
}

// Search searches for the most similar vectors.
func (s *SimpleStore) Search(vector []float32, limit int, threshold float32) ([]Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Result
	now := time.Now()

	for _, entry := range s.entries {
		if !entry.ExpirationTime.IsZero() && entry.ExpirationTime.Before(now) {
			continue // Expired
		}

		if len(entry.Vector) != len(vector) {
			continue
		}
		sim := cosineSimilarity(vector, entry.Vector)
		if sim >= threshold {
			results = append(results, Result{
				Data:       entry.Data,
				Similarity: sim,
			})
		}
	}

	// Sort by similarity descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Clear clears the store.
func (s *SimpleStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make([]Entry, 0, s.maxCapacity)
}

func cosineSimilarity(a, b []float32) float32 {
	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}
