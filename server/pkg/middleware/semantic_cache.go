// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary: Interface for generating embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
	// Summary: Generates an embedding for text.
	//
	// Parameters:
	//   - ctx: context.Context. The context.
	//   - text: string. The input text.
	//
	// Returns:
	//   - []float32: The embedding vector.
	//   - error: An error if generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary: Interface for vector storage operations.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Summary: Stores a vector and its result.
	//
	// Parameters:
	//   - ctx: context.Context. The context.
	//   - key: string. The cache key.
	//   - vector: []float32. The embedding vector.
	//   - result: any. The cached result.
	//   - ttl: time.Duration. Time to live.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Summary: Searches for a similar vector.
	//
	// Parameters:
	//   - ctx: context.Context. The context.
	//   - key: string. The cache key.
	//   - query: []float32. The query vector.
	//
	// Returns:
	//   - any: The found result.
	//   - float32: The similarity score.
	//   - bool: True if a match was found.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Summary: Removes expired entries.
	//
	// Parameters:
	//   - ctx: context.Context. The context.
	//   - key: string. The cache key.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Summary: Cache implementation using semantic similarity.
//
// Fields:
//   - provider: EmbeddingProvider. Provider for embeddings.
//   - store: VectorStore. Storage for vectors.
//   - threshold: float32. Similarity threshold.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// Summary: Initializes a new SemanticCache.
//
// Parameters:
//   - provider: EmbeddingProvider. The embedding provider.
//   - store: VectorStore. The vector store.
//   - threshold: float32. The similarity threshold.
//
// Returns:
//   - *SemanticCache: The initialized cache.
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
// Summary: Retrieves a cached result based on semantic similarity.
//
// Parameters:
//   - ctx: context.Context. The context.
//   - key: string. The cache key.
//   - input: string. The input text to match.
//
// Returns:
//   - any: The cached result (if found).
//   - []float32: The computed embedding of the input.
//   - bool: True if a hit was found.
//   - error: An error if retrieval fails.
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
// Summary: Caches a result with its embedding.
//
// Parameters:
//   - ctx: context.Context. The context.
//   - key: string. The cache key.
//   - embedding: []float32. The embedding vector.
//   - result: any. The result to cache.
//   - ttl: time.Duration. Time to live.
//
// Returns:
//   - error: An error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
