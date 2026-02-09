// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
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
	// ctx is the context for the request.
	// key is the key.
	// vector is the vector.
	// result is the result.
	// ttl is the ttl.
	//
	// Returns an error if the operation fails.
	Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error
	// Search searches for the most similar entry in the vector store.
	//
	// ctx is the context for the request.
	// key is the key.
	// query is the query.
	//
	// Returns the result.
	// Returns the result.
	// Returns true if successful.
	Search(ctx context.Context, key string, query []float32) (any, float32, bool)
	// Prune removes expired entries.
	//
	// ctx is the context for the request.
	// key is the key.
	Prune(ctx context.Context, key string)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
type SemanticCache struct {
	provider   EmbeddingProvider
	store      VectorStore
	threshold  float32
	exactCache *gocache.Cache
}

// NewSemanticCache creates a new SemanticCache.
//
// provider is the provider.
// store is the store.
// threshold is the threshold.
//
// Returns the result.
func NewSemanticCache(provider EmbeddingProvider, store VectorStore, threshold float32) *SemanticCache {
	if threshold <= 0 {
		threshold = 0.9 // Default high threshold
	}
	if store == nil {
		store = NewSimpleVectorStore()
	}
	return &SemanticCache{
		provider:   provider,
		store:      store,
		threshold:  threshold,
		exactCache: gocache.New(5*time.Minute, 10*time.Minute),
	}
}

// Get attempts to find a semantically similar cached result.
// It returns the result, the computed embedding, a boolean indicating a hit, and an error.
func (c *SemanticCache) Get(ctx context.Context, key string, input string) (any, []float32, bool, error) {
	// ⚡ BOLT: Added L1 exact match cache to skip expensive embedding generation for repeated queries.
	// Randomized Selection from Top 5 High-Impact Targets
	// Use composite key to prevent cross-tool collisions
	cacheKey := key + ":" + input
	if val, found := c.exactCache.Get(cacheKey); found {
		return val, nil, true, nil
	}

	embedding, err := c.provider.Embed(ctx, input)
	if err != nil {
		return nil, nil, false, err
	}

	result, score, found := c.store.Search(ctx, key, embedding)
	if found && score >= c.threshold {
		// Populate L1 cache on semantic hit for faster subsequent access
		c.exactCache.Set(cacheKey, result, gocache.DefaultExpiration)
		return result, embedding, true, nil
	}
	return nil, embedding, false, nil
}

// Set adds a result to the cache using the provided embedding.
//
// ctx is the context for the request.
// key is the key.
// input is the input string that generated the result.
// embedding is the embedding.
// result is the result.
// ttl is the ttl.
//
// Returns an error if the operation fails.
func (c *SemanticCache) Set(ctx context.Context, key string, input string, embedding []float32, result any, ttl time.Duration) error {
	// Populate L1 cache
	c.exactCache.Set(key+":"+input, result, ttl)
	return c.store.Add(ctx, key, embedding, result, ttl)
}
