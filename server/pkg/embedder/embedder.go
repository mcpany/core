// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package embedder

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strings"
	"unicode"
)

// Embedder defines the interface for converting text to vector embeddings.
type Embedder interface {
	Embed(text string) ([]float32, error)
	Dimension() int
}

// BagOfWordsEmbedder implements a simple hashing-based bag-of-words embedder.
// It maps words to a fixed-size vector space using hashing.
type BagOfWordsEmbedder struct {
	dimension int
}

// NewBagOfWordsEmbedder creates a new BagOfWordsEmbedder with the given dimension.
func NewBagOfWordsEmbedder(dimension int) *BagOfWordsEmbedder {
	return &BagOfWordsEmbedder{
		dimension: dimension,
	}
}

// Dimension returns the dimension of the embeddings.
func (e *BagOfWordsEmbedder) Dimension() int {
	return e.dimension
}

// Embed computes the embedding for the given text.
func (e *BagOfWordsEmbedder) Embed(text string) ([]float32, error) {
	if text == "" {
		return make([]float32, e.dimension), nil
	}

	// Tokenize
	tokens := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(tokens) == 0 {
		return make([]float32, e.dimension), nil
	}

	vector := make([]float32, e.dimension)

	for _, token := range tokens {
		// Hash token to an index
		hash := sha256.Sum256([]byte(token))
		// Use first 8 bytes as uint64 to determine index
		val := binary.BigEndian.Uint64(hash[:8])
		index := int(val % uint64(e.dimension))

		vector[index] += 1.0
	}

	// Normalize
	return normalize(vector), nil
}

func normalize(v []float32) []float32 {
	var sumSq float32
	for _, val := range v {
		sumSq += val * val
	}
	if sumSq == 0 {
		return v
	}
	norm := float32(math.Sqrt(float64(sumSq)))
	for i := range v {
		v[i] /= norm
	}
	return v
}
