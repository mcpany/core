// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package llm provides interfaces and implementations for interacting with LLM providers.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the interface for an LLM client.
//
// Summary: is the interface for an LLM client.
type Client interface {
	// ChatCompletion sends a chat request to the LLM and returns the response.
	//
	// Summary:
	//   Sends a chat completion request to the configured LLM provider.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - req: ChatRequest. The chat request parameters.
	//
	// Returns:
	//   - *ChatResponse: The chat response from the LLM.
	//   - error: An error if the request fails or the response is invalid.
	//
	// Errors:
	//   - Returns error if marshaling fails, network request fails, or API returns non-200 status.
	//
	// Side Effects:
	//   - Makes a network request to the LLM provider.
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ChatRequest represents a chat completion request.
//
// Summary: represents a chat completion request.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message.
//
// Summary: represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a chat completion response.
//
// Summary: represents a chat completion response.
type ChatResponse struct {
	Content string `json:"content"`
}

// OpenAIClient implements Client for OpenAI.
//
// Summary: implements Client for OpenAI.
type OpenAIClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIClient creates a new OpenAIClient.
//
// Summary: creates a new OpenAIClient.
//
// Parameters:
//   - apiKey: string. The apiKey.
//   - baseURL: string. The baseURL.
//
// Returns:
//   - *OpenAIClient: The *OpenAIClient.
func NewOpenAIClient(apiKey string, baseURL string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type openAIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ChatCompletion performs a chat completion request.
//
// Summary: performs a chat completion request.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: ChatRequest. The req.
//
// Returns:
//   - *ChatResponse: The *ChatResponse.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (c *OpenAIClient) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	reqBody := openAIChatRequest(req)
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("openai api error (status %d): failed to read body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(body))
	}

	var response openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("openai error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	return &ChatResponse{
		Content: response.Choices[0].Message.Content,
	}, nil
}
