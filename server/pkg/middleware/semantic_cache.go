// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"
)

// EmbeddingProvider represents the public EmbeddingProvider entity.
//
// Summary: Defines the structured data model representing a provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// VectorStore represents the public VectorStore entity.
//
// Summary: Defines the structured data model representing a store.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// SemanticCache represents the public SemanticCache entity.
//
// Summary: Defines the structured data model representing a cache.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     VectorStore
	threshold float32
}

// NewSemanticCache serves as a public interface for interacting with NewSemanticCache.
//
// Summary: Constructs and returns an initialized semantic cache ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Get serves as a public interface for interacting with Get.
//
// Summary: Fetches and returns the underlying  from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Set serves as a public interface for interacting with Set.
//
// Summary: Set the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
