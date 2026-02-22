// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary: Interface for generating vector embeddings from text.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
	// Summary: Generates an embedding vector for the given text.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - text: string. The text input to embed.
	//
	// Returns:
	//   - []float32: The embedding vector.
	//   - error: An error if the embedding generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary: Interface for a vector storage backend.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Summary: Adds a vector and its associated result to the store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. A grouping key (e.g. tool name).
	//   - vector: []float32. The embedding vector.
	//   - result: any. The tool execution result to cache.
	//   - ttl: time.Duration. Time-to-live for the cache entry.
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
	//   - key: string. A grouping key (e.g. tool name).
	//   - query: []float32. The query vector.
	//
	// Returns:
	//   - any: The cached result.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a match was found.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Summary: Removes expired entries from the store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. A grouping key (e.g. tool name).
	//
	// Returns:
	//   - None.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Summary: A cache that uses semantic similarity to retrieve results.
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
//   - provider: EmbeddingProvider. The provider to use for generating embeddings.
//   - store: VectorStore. The storage backend for vectors.
//   - threshold: float32. The similarity threshold (0.0 to 1.0) for cache hits. Defaults to 0.9.
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
// Summary: Retrieves a cached result if a semantically similar query is found.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - key: string. A grouping key (e.g. tool name).
//   - input: string. The input text to search for.
//
// Returns:
//   - any: The cached result (if found).
//   - []float32: The generated embedding for the input (useful for subsequent Set).
//   - bool: True if a valid cache hit occurred.
//   - error: An error if embedding generation fails.
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
//   - key: string. A grouping key (e.g. tool name).
//   - embedding: []float32. The embedding vector (returned from Get).
//   - result: any. The result to cache.
//   - ttl: time.Duration. Time-to-live for the cache entry.
//
// Returns:
//   - error: An error if the add operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
