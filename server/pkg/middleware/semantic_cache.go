// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

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
	// Prune removes expired entries.
	//
	// ctx is the context for the request.
	// key is the key.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// provider is the provider.
// store is the store.
// threshold is the threshold.
//
// Returns the result.
func NewSemanticCache(provider EmbeddingProvider, store VectorStore, threshold float32) *SemanticCache {
	if threshold <= 0 {
		threshold = 0.9 // Default high threshold
	}
	if store == nil {
		store = NewSimpleVectorStore()
	}
	return &SemanticCache{
		provider:  provider,
		store:     store,
		threshold: threshold,
	}
}

// Get attempts to find a semantically similar cached result.
// It returns the result, the computed embedding, a boolean indicating a hit, and an error.
func (c *SemanticCache) Get(ctx context.Context, key string, input string) (any, []float32, bool, error) {
	embedding, err := c.provider.Embed(ctx, input)
	if err != nil {
		return nil, nil, false, err
	}

	result, score, found := c.store.Search(ctx, key, embedding)
	if found && score >= c.threshold {
		return result, embedding, true, nil
	}
	return nil, embedding, false, nil
}

// Set adds a result to the cache using the provided embedding.
//
// ctx is the context for the request.
// key is the key.
// embedding is the embedding.
// result is the result.
// ttl is the ttl.
//
// Returns an error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
