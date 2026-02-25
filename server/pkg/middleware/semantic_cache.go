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
	// Summary: Generates an embedding vector.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - text (string): The text to embed.
	//
	// Returns:
	//   - []float32: The embedding vector.
	//   - error: An error if the embedding generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary: Interface for vector storage and retrieval.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Summary: Stores a vector and its associated result.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The key associated with the entry (e.g., service ID).
	//   - vector ([]float32): The vector to store.
	//   - result (any): The result object to cache.
	//   - ttl (time.Duration): The time-to-live for the entry.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Summary: Finds the nearest neighbor for a query vector.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The key to filter search results by.
	//   - query ([]float32): The query vector.
	//
	// Returns:
	//   - any: The cached result object.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a match was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Summary: Removes expired entries from the store.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): Optional key to filter pruning (empty for all).
	//
	// Side Effects:
	//   - Deletes data from the underlying storage.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Summary: Caching middleware that uses semantic similarity.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// Summary: Initializes a new SemanticCache instance.
//
// Parameters:
//   - provider (EmbeddingProvider): The provider for generating embeddings.
//   - store (VectorStore): The storage backend for vectors.
//   - threshold (float32): The minimum similarity score required for a cache hit.
//
// Returns:
//   - *SemanticCache: A pointer to the new SemanticCache.
//
// Side Effects:
//   - Allocates memory for the cache structure.
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
//   - key (string): The key to identify the cache namespace (e.g. service ID).
//   - input (string): The input text to search for.
//
// Returns:
//   - any: The cached result if found.
//   - []float32: The computed embedding for the input (can be reused for Set).
//   - bool: True if a cache hit occurred, false otherwise.
//   - error: An error if embedding generation fails.
//
// Side Effects:
//   - Calls the embedding provider (network call).
//   - Queries the vector store (database/memory read).
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
//   - key (string): The key to identify the cache namespace.
//   - embedding ([]float32): The embedding vector for the input.
//   - result (any): The result to cache.
//   - ttl (time.Duration): The expiration time for the cache entry.
//
// Returns:
//   - error: An error if the storage operation fails.
//
// Side Effects:
//   - Writes to the vector store.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
