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
	// Summary: Stores a vector embedding and its associated result with an expiration time.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - key: string. The unique key for grouping entries (e.g. tool name).
	//   - vector: []float32. The embedding vector.
	//   - result: any. The result object to store.
	//   - ttl: time.Duration. The time-to-live for the entry.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Summary: Finds the nearest neighbor vector that exceeds the similarity threshold.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - key: string. The unique key to search within.
	//   - query: []float32. The query vector.
	//
	// Returns:
	//   - any: The cached result.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a match was found.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
	// Summary: Cleans up entries that have exceeded their TTL.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - key: string. The unique key to prune (optional, if supported by store).
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
// Summary: Initializes a new SemanticCache with an embedding provider and a vector store.
//
// Parameters:
//   - provider: EmbeddingProvider. The provider used to generate embeddings for inputs.
//   - store: VectorStore. The backend used to store and search vectors.
//   - threshold: float32. The minimum cosine similarity score (0.0-1.0) required for a cache hit.
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
// Summary: Generates an embedding for the input and searches the vector store for a similar previous result.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - key: string. The cache partition key (e.g. tool name).
//   - input: string. The input text to match against.
//
// Returns:
//   - any: The cached result (if hit).
//   - []float32: The computed embedding vector (useful for setting cache on miss).
//   - bool: True if a cache hit occurred.
//   - error: An error if embedding generation or search fails.
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
// Summary: Stores a result in the semantic cache associated with the given embedding vector.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - key: string. The cache partition key.
//   - embedding: []float32. The vector embedding of the input.
//   - result: any. The result to cache.
//   - ttl: time.Duration. The expiration duration.
//
// Returns:
//   - error: An error if the storage operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
