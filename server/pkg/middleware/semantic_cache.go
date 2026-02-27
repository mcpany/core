// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/mcpany/core/server/pkg/vector"
)

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
type SemanticCache struct {
	provider  vector.EmbeddingProvider
	store     vector.VectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
//
// provider is the provider.
// store is the store.
// threshold is the threshold.
//
// Returns the result.
func NewSemanticCache(provider vector.EmbeddingProvider, store vector.VectorStore, threshold float32) *SemanticCache {
	if threshold <= 0 {
		threshold = 0.9 // Default high threshold
	}
	if store == nil {
		store = vector.NewSimpleVectorStore()
	}
	return &SemanticCache{
		provider:  provider,
		store:     store,
		threshold: threshold,
	}
}

// Get attempts to find a semantically similar cached result.
// It returns the result, the computed embedding, a boolean indicating a hit, and an error.
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
// ctx is the context for the request.
// key is the key.
// embedding is the embedding.
// result is the result.
// ttl is the ttl.
//
// Returns an error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
