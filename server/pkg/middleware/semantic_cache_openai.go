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

// OpenAIEmbeddingProvider represents the public OpenAIEmbeddingProvider entity.
//
// Summary: Defines the structured data model representing a ai embedding provider.
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
type OpenAIEmbeddingProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAIEmbeddingProvider serves as a public interface for interacting with NewOpenAIEmbeddingProvider.
//
// Summary: Constructs and returns an initialized open ai embedding provider ready for consumption.
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
func NewOpenAIEmbeddingProvider(apiKey, model string) *OpenAIEmbeddingProvider {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &OpenAIEmbeddingProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1/embeddings",
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type openAIEmbeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed serves as a public interface for interacting with Embed.
//
// Summary: Embed the  appropriately based on current system conditions.
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
func (p *OpenAIEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := openAIEmbeddingRequest{
		Input:          text,
		Model:          p.model,
		EncodingFormat: "float",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(body))
	}

	var response openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("openai error: %s", response.Error.Message)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return response.Data[0].Embedding, nil
}
