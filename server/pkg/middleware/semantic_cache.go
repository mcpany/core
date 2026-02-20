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
	//  ctx (context.Context): The context for the request.
	//  text (string): The text input to embed.
	//
	// Returns:
	//  []float32: The embedding vector.
	//  error: An error if the embedding generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
	// Parameters:
	//  ctx (context.Context): The context for the request.
	//  key (string): A key to scope the vector (e.g. tool name).
	//  vector ([]float32): The vector to store.
	//  result (any): The data associated with the vector.
	//  ttl (time.Duration): The time-to-live for the entry.
	//
	// Returns:
	//  error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
	// Parameters:
	//  ctx (context.Context): The context for the request.
	//  key (string): The scope key to search within.
	//  query ([]float32): The query vector.
	//
	// Returns:
	//  any: The associated result if found.
	//  float32: The similarity score (0.0 to 1.0).
	//  bool: True if a result was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries from the store.
	//
	// Parameters:
	//  ctx (context.Context): The context for the request.
	//  key (string): The scope key to prune.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
// It stores results indexed by the vector representation of the input.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// Parameters:
//  provider (EmbeddingProvider): The provider to generate embeddings.
//  store (VectorStore): The store to save and search vectors.
//  threshold (float32): The minimum similarity score (0-1) to consider a hit. Defaults to 0.9.
//
// Returns:
//  *SemanticCache: The initialized semantic cache.
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

// Get attempts to find a semantically similar cached result for the given input.
//
// Parameters:
//  ctx (context.Context): The context for the request.
//  key (string): The scope key (e.g. tool name).
//  input (string): The text input to search for.
//
// Returns:
//  any: The cached result if found.
//  []float32: The embedding of the input (useful for Set if cache miss).
//  bool: True if a cache hit occurred.
//  error: An error if embedding generation fails.
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
//  ctx (context.Context): The context for the request.
//  key (string): The scope key.
//  embedding ([]float32): The vector representation of the input.
//  result (any): The result to cache.
//  ttl (time.Duration): The time-to-live for the cached item.
//
// Returns:
//  error: An error if the add operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
