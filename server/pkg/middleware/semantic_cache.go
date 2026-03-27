// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary. Interface for services that can generate vector embeddings from text.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	//
// Parameters.
	//   - ctx: context.Context. The request context.
	//   - text: string. The text to embed.
	//
// Returns.
	//   - []float32: The resulting embedding vector.
	//   - error: An error if generation fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorStore defines the interface for storing and searching vectors.
//
// Summary. Interface for storage backends that support vector similarity search.
type VectorStore interface {
	// Add adds a new entry to the vector store.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The unique key for the entry.
	//   - vector: []float32. The embedding vector.
	//   - result: any. The associated result data.
	//   - ttl: time.Duration. The time-to-live for the entry.
	//
// Returns.
	//   - error: An error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error

	// Search searches for the most similar entry in the vector store.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//   - key: string. The key to restrict the search scope.
	//   - query: []float32. The query embedding vector.
	//
// Returns.
	//   - any: The best matching result data.
	//   - float32: The similarity score (0.0 to 1.0).
	//   - bool: True if a match was found, false otherwise.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)

	// Prune removes expired entries.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//   - key: string. Optional key to restrict pruning scope.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
//
// Summary. A cache implementation that uses semantic similarity rather than exact key matching.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache provides newsemanticcache functionality.
//
// Summary: NewSemanticCache.
//
// Parameters.
//   - provider: The parameter.
//   - store: The parameter.
//   - threshold: The parameter.
//
// Returns.
//   - result: The result.
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

// Get provides get functionality.
//
// Summary: Get.
//
// Parameters.
//   - ctx: The parameter.
//   - key: The parameter.
//   - input: The parameter.
//   - []float32: The parameter.
//   - bool: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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

// Set provides set functionality.
//
// Summary: Set.
//
// Parameters.
//   - ctx: The parameter.
//   - key: The parameter.
//   - embedding: The parameter.
//   - result: The parameter.
//   - ttl: The parameter.
//
// Returns.
//   - result: The result.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
