// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary: Interface for generating text embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
	// Summary: Generates an embedding vector for the input text.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - text: string. The text to embed.
	//
	// Returns:
	//   - []float32: The embedding vector.
	//   - error: An error if embedding generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary: Interface for vector storage operations.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Summary: Adds a vector entry to the store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The partition key (e.g., user ID).
	//   - vector: []float32. The vector to store.
	//   - result: any. The associated result/data.
	//   - ttl: time.Duration. Time-to-live for the entry.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Summary: Searches for the nearest neighbor vector.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The partition key.
	//   - query: []float32. The query vector.
	//
	// Returns:
	//   - any: The result/data associated with the best match.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a match was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Summary: Removes expired entries from the store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The partition key.
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
// Summary: Initializes a new SemanticCache middleware.
//
// Parameters:
//   - provider: EmbeddingProvider. The provider for text embeddings.
//   - store: VectorStore. The storage backend for vectors.
//   - threshold: float32. The similarity threshold for a cache hit (0.0 to 1.0).
//
// Returns:
//   - *SemanticCache: The initialized semantic cache.
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
//   - ctx: context.Context. The context for the request.
//   - key: string. The partition key.
//   - input: string. The input text to search for.
//
// Returns:
//   - any: The cached result if found.
//   - []float32: The embedding of the input text.
//   - bool: True if a cache hit occurred.
//   - error: An error if the operation fails.
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
// Summary: Caches a result with its embedding vector.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - key: string. The partition key.
//   - embedding: []float32. The embedding vector of the input.
//   - result: any. The result to cache.
//   - ttl: time.Duration. The time-to-live for the cache entry.
//
// Returns:
//   - error: An error if storing fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
