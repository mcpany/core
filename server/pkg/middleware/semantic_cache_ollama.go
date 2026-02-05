// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaEmbeddingProvider implements EmbeddingProvider for Ollama.
type OllamaEmbeddingProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbeddingProvider creates a new OllamaEmbeddingProvider.
//
// Summary: Initializes an embedding provider that connects to an Ollama server.
//
// Parameters:
//   - baseURL: string. The base URL of the Ollama API (e.g., "http://localhost:11434").
//   - model: string. The name of the model to use for embeddings (e.g., "nomic-embed-text").
//
// Returns:
//   - *OllamaEmbeddingProvider: The initialized provider.
func NewOllamaEmbeddingProvider(baseURL, model string) *OllamaEmbeddingProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return &OllamaEmbeddingProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// Embed generates an embedding for the given text using Ollama.
//
// Summary: Calls the Ollama API to generate a vector embedding for the input text.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - text: string. The input text to embed.
//
// Returns:
//   - []float32: The generated embedding vector.
//   - error: An error if the API call fails.
func (p *OllamaEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := ollamaEmbeddingRequest{
		Model:  p.model,
		Prompt: text,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama api error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return response.Embedding, nil
}
