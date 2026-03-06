// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider - Auto-generated documentation.
//
// Summary: EmbeddingProvider defines the interface for fetching text embeddings.
//
// Methods:
//   - Various methods for EmbeddingProvider.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
	// Parameters:
	//   - ctx: context.Context. The request context.
	//   - text: string. The text to embed.
	//
	// Returns:
	//   - []float32: The resulting embedding vector.
	//   - error: An error if generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore - Auto-generated documentation.
//
// Summary: VectorStore defines the interface for storing and searching vectors.
//
// Methods:
//   - Various methods for VectorStore.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The unique key for the entry.
	//   - vector: []float32. The embedding vector.
	//   - result: any. The associated result data.
	//   - ttl: time.Duration. The time-to-live for the entry.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The key to restrict the search scope.
	//   - query: []float32. The query embedding vector.
	//
	// Returns:
	//   - any: The best matching result data.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a match was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - key: string. Optional key to restrict pruning scope.
	Prune(ctx context.Context, key string)
}

// SemanticCache - Auto-generated documentation.
//
// Summary: SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Fields:
//   - Various fields for SemanticCache.
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
//   - provider: EmbeddingProvider. The service to generate embeddings.
//   - store: VectorStore. The storage backend for vectors.
//   - threshold: float32. The minimum similarity score (0-1) to consider a hit.
//
// Returns:
//   - *SemanticCache: The initialized semantic cache.
//
// Side Effects:
//   - Sets a default threshold of 0.9 if the provided threshold is <= 0.
//   - Creates a memory-based vector store if store is nil.
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
// Summary: Retrieves a cached result if a semantically similar entry exists.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - key: string. The semantic key or scope.
//   - input: string. The query text to match against.
//
// Returns:
//   - any: The cached result if found.
//   - []float32: The embedding generated for the input text (useful for subsequent Set).
//   - bool: True if a cache hit occurred.
//   - error: An error if embedding generation fails.
//
// Errors:
//   - Returns error if the embedding provider fails.
//
// Side Effects:
//   - calls the EmbeddingProvider to generate an embedding.
//   - calls the VectorStore to search for matches.
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
// Summary: Caches a result associated with a specific embedding.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - key: string. The semantic key or scope.
//   - embedding: []float32. The embedding vector (usually returned from Get).
//   - result: any. The result data to cache.
//   - ttl: time.Duration. The expiration time for the cache entry.
//
// Returns:
//   - error: An error if the storage operation fails.
//
// Side Effects:
//   - Writes to the underlying VectorStore.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
