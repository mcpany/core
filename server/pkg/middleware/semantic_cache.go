// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
// EmbeddingProvider defines the interface for fetching text embeddings.
//
// Summary: Interface for services that can generate vector embeddings from text.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
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
// VectorStore defines the interface for storing and searching vectors.
//
// Summary: Interface for storage backends that support vector similarity search.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
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
	// Prune removes expired entries.
// NewSemanticCache creates a new SemanticCache.
//
// Summary: Initializes a new SemanticCache.
//
// Parameters:
//   - provider: EmbeddingProvider. The service to generate embeddings.
//   - store: VectorStore. The storage backend for vectors.
// Get attempts to find a semantically similar cached result.
//
// Summary: Retrieves a cached result if a semantically similar entry exists.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - key: string. The semantic key or scope.
//   - input: string. The query text to match against.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - any: The cached result if found.
//   - []float32: The embedding generated for the input text (useful for subsequent Set).
//   - bool: True if a cache hit occurred.
//   - error: An error if embedding generation fails.
//
// Errors:
//   - Returns error if the embedding provider fails.
// Set adds a result to the cache using the provided embedding.
//
// Summary: Caches a result associated with a specific embedding.
//
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// Errors:
//   - triggers relevant error states on failure.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	return c.store.Add(ctx, key, embedding, result, ttl)
}
