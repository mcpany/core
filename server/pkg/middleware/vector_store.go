// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// VectorStore defines the interface for storing and retrieving vectors.
type VectorStore interface {
	// Add adds a vector and its associated result to the store.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error
	// Search finds the nearest neighbor for the given query vector.
	// It returns the result, the score (cosine similarity), and a boolean indicating if a match was found.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool, error)
	// Close closes the store and releases resources.
	Close() error
}
