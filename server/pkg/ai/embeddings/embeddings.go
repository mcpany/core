// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package embeddings

import (
	"context"
)

// Embedding represents a vector embedding.
type Embedding []float32

// Embedder is the interface for generating embeddings.
type Embedder interface {
	// Embed generates embeddings for the given texts.
	Embed(ctx context.Context, texts []string) ([]Embedding, error)
}
