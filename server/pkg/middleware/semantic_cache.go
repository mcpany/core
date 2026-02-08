// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary: defines the interface for fetching text embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
	// Summary: generates an embedding vector for the given text.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - text: string. The string.
	//
	// Returns:
	//   - []float32: The []float32.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary: defines the interface for storing and searching vectors.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Summary: adds a new entry to the vector store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - key: string. The string.
	//   - vector: []float32. The []float32.
	//   - result: any. The any.
	//   - ttl: time.Duration. The duration.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error
	// Search searches for the most similar entry in the vector store.
	//
	// Summary: searches for the most similar entry in the vector store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - key: string. The string.
	//   - query: []float32. The []float32.
	//
	// Returns:
	//   - any: The any.
	//   - float32: The float32.
	//   - bool: The bool.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)
	// Prune removes expired entries.
	//
	// Summary: removes expired entries.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - key: string. The string.
	//
	// Returns:
	//   None.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Summary: implements a semantic cache using embeddings and cosine similarity.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// Summary: creates a new SemanticCache.
//
// Parameters:
//   - provider: EmbeddingProvider. The provider.
//   - store: VectorStore. The store.
//   - threshold: float32. The threshold.
//
// Returns:
//   - *SemanticCache: The *SemanticCache.
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
//
// Summary: attempts to find a semantically similar cached result.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - key: string. The key.
//   - input: string. The input.
//
// Returns:
//   - any: The any.
//   - []float32: The []float32.
//   - bool: The bool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: adds a result to the cache using the provided embedding.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - key: string. The key.
//   - embedding: []float32. The embedding.
//   - result: any. The result.
//   - ttl: time.Duration. The ttl.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
