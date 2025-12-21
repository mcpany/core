// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package embeddings

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
)

const VectorDimension = 1024

// LocalEmbedder implements a simple N-Gram hashing embedder.
type LocalEmbedder struct{}

// NewLocalEmbedder creates a new LocalEmbedder.
func NewLocalEmbedder() *LocalEmbedder {
	return &LocalEmbedder{}
}

// Embed generates embeddings for the given texts using hashing trick.
func (e *LocalEmbedder) Embed(ctx context.Context, texts []string) ([]Embedding, error) {
	result := make([]Embedding, len(texts))
	for i, text := range texts {
		result[i] = e.computeEmbedding(text)
	}
	return result, nil
}

func (e *LocalEmbedder) computeEmbedding(text string) Embedding {
	vector := make(Embedding, VectorDimension)
	if text == "" {
		return vector
	}

	// Simple trigram hashing
	normalized := strings.ToLower(text)
	runes := []rune(normalized)

	if len(runes) < 3 {
		// Fallback for short strings
		h := fnv.New32a()
		_, _ = h.Write([]byte(normalized))
		idx := h.Sum32() % VectorDimension
		vector[idx] = 1.0
	} else {
		for i := 0; i <= len(runes)-3; i++ {
			trigram := string(runes[i : i+3])
			h := fnv.New32a()
			_, _ = h.Write([]byte(trigram))
			idx := h.Sum32() % VectorDimension
			vector[idx] += 1.0
		}
	}

	// Normalize
	var sumSq float64
	for _, v := range vector {
		sumSq += float64(v * v)
	}
	magnitude := float32(math.Sqrt(sumSq))

	if magnitude > 0 {
		for i := range vector {
			vector[i] /= magnitude
		}
	}

	return vector
}
