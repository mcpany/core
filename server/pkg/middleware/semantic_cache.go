// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary: Interface for text embedding generation.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - text (string): The text to embed.
	//
	// Returns:
	//   - ([]float32): The generated embedding vector.
	//   - (error): An error if the embedding generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary: Interface for vector storage and retrieval.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The unique key for the entry.
	//   - vector ([]float32): The vector to store.
	//   - result (any): The result associated with the vector.
	//   - ttl (time.Duration): The time-to-live for the entry.
	//
	// Returns:
	//   - (error): An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The key to search within.
	//   - query ([]float32): The query vector.
	//
	// Returns:
	//   - (any): The most similar result found.
	//   - (float32): The similarity score (0.0 to 1.0).
	//   - (bool): True if a match was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The key to prune.
	//
	// Side Effects:
	//   - Removes expired entries from the store.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Summary: Semantic caching implementation.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// Summary: Initializes a new semantic cache.
//
// Parameters:
//   - provider (EmbeddingProvider): The embedding provider to use.
//   - store (VectorStore): The vector store to use.
//   - threshold (float32): The similarity threshold for cache hits (0.0 to 1.0).
//
// Returns:
//   - (*SemanticCache): The initialized semantic cache.
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
//   - ctx (context.Context): The context for the request.
//   - key (string): The cache key scope.
//   - input (string): The input text to search for.
//
// Returns:
//   - (any): The cached result if found.
//   - ([]float32): The embedding of the input text (useful for subsequent Set).
//   - (bool): True if a cache hit occurred, false otherwise.
//   - (error): An error if embedding generation or search fails.
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
//   - ctx (context.Context): The context for the request.
//   - key (string): The cache key scope.
//   - embedding ([]float32): The embedding vector for the input.
//   - result (any): The result to cache.
//   - ttl (time.Duration): The time-to-live for the cache entry.
//
// Returns:
//   - (error): An error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
