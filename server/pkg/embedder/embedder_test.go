// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package embedder_test

import (
	"testing"

	"github.com/mcpany/core/pkg/embedder"
	"github.com/stretchr/testify/assert"
)

func TestBagOfWordsEmbedder(t *testing.T) {
	e := embedder.NewBagOfWordsEmbedder(10)

	t.Run("Empty string", func(t *testing.T) {
		v, err := e.Embed("")
		assert.NoError(t, err)
		assert.Equal(t, 10, len(v))
		for _, val := range v {
			assert.Equal(t, float32(0), val)
		}
	})

	t.Run("Deterministic", func(t *testing.T) {
		v1, _ := e.Embed("hello world")
		v2, _ := e.Embed("hello world")
		assert.Equal(t, v1, v2)
	})

	t.Run("Similarity", func(t *testing.T) {
		// "apple banana" should be closer to "apple orange" than "car truck"
		// Use larger dim to minimize collisions
		e2 := embedder.NewBagOfWordsEmbedder(100)
		vBase, _ := e2.Embed("apple banana")
		vSim, _ := e2.Embed("apple orange")
		vDiff, _ := e2.Embed("car truck")

		sim1 := dot(vBase, vSim) // Since normalized, dot product is cosine sim
		sim2 := dot(vBase, vDiff)

		assert.True(t, sim1 > sim2, "Expected 'apple banana' to be closer to 'apple orange' than 'car truck'")
	})
}

func dot(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
