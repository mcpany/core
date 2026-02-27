// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	// It returns the embedding as a slice of float32 and any error encountered.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// ctx is the context for the request.
	// key is the key.
	// vector is the vector.
	// result is the result.
	// ttl is the ttl.
	//
	// Returns an error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error
	// Search searches for the most similar entry in the vector store.
	//
	// ctx is the context for the request.
	// key is the key.
	// query is the query.
	//
	// Returns the result.
	// Returns the result.
	// Returns true if successful.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)
	// SearchTopK searches for the top k most similar entries in the vector store.
	//
	// ctx is the context for the request.
	// key is the key.
	// query is the query.
	// k is the maximum number of results to return.
	//
	// Returns:
	//   - []any: A slice of results.
	//   - []float32: A slice of similarity scores.
	//   - error: An error if the search fails.
	SearchTopK(ctx context.Context, key string, query []float32, k int) ([]any, []float32, error)
	// Prune removes expired entries.
	//
	// ctx is the context for the request.
	// key is the key.
	Prune(ctx context.Context, key string)
}
