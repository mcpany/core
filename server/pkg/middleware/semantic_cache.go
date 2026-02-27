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
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - text (string): The text input to generate an embedding for.
	//
	// Returns:
	//   - []float32: The embedding vector.
	//   - error: An error if the embedding generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The unique key identifying the entry (e.g., tool name).
	//   - vector ([]float32): The embedding vector associated with the entry.
	//   - result (any): The result object to store.
	//   - ttl (time.Duration): The time-to-live for the entry.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): The key to scope the search (e.g., tool name).
	//   - query ([]float32): The query embedding vector.
	//
	// Returns:
	//   - any: The stored result of the most similar entry.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a matching entry was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries from the store.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - key (string): Optional key to scope the pruning. If empty, prunes all expired entries.
	//
	// Side Effects:
	//   - Deletes expired entries from the underlying storage.
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
// Parameters:
//   - provider (EmbeddingProvider): The provider to use for generating embeddings.
//   - store (VectorStore): The backing store for vectors and results.
//   - threshold (float32): The similarity threshold (0.0-1.0) for a cache hit. Defaults to 0.9 if <= 0.
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
// It generates an embedding for the input and searches the vector store.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - key (string): The key to scope the cache lookup (e.g., tool name).
//   - input (string): The input text to match against the cache.
//
// Returns:
//   - any: The cached result if a hit occurs.
//   - []float32: The embedding generated for the input (useful for subsequent Set calls).
//   - bool: True if a cache hit occurred (similarity >= threshold).
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
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - key (string): The key to associate with the entry.
//   - embedding ([]float32): The embedding vector for the input.
//   - result (any): The result to cache.
//   - ttl (time.Duration): The time-to-live for the cached entry.
//
// Returns:
//   - error: An error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
