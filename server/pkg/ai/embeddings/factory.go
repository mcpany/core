// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package embeddings

import (
	"fmt"
)

// NewEmbedder creates a new Embedder based on the provider.
func NewEmbedder(provider string, model string) (Embedder, error) {
	switch provider {
	case "local":
		return NewLocalEmbedder(), nil
	case "openai":
		return nil, fmt.Errorf("openai provider not implemented yet")
	default:
		// Default to local for now
		return NewLocalEmbedder(), nil
	}
}
