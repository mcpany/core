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

// Client defines the interface for an LLM (Large Language Model) client.
// Implementations of this interface allow interacting with different LLM providers
// (e.g., OpenAI, Gemini, Anthropic) in a uniform way.
type Client interface {
	// ChatCompletion sends a chat request to the LLM and returns the response.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - req: ChatRequest. The chat completion request parameters.
	//
	// Returns:
	//   - *ChatResponse: The chat completion response.
	//   - error: An error if the request fails.
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ChatRequest represents a standard chat completion request structure.
type ChatRequest struct {
	// Model specifies the identifier of the LLM model to use.
	Model string `json:"model"`
	// Messages is a list of conversation messages.
	Messages []Message `json:"messages"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	// Role indicates the sender of the message (e.g., "user", "assistant", "system").
	Role string `json:"role"`
	// Content is the text content of the message.
	Content string `json:"content"`
}

// ChatResponse represents the response from a chat completion request.
type ChatResponse struct {
	// Content is the generated text content from the LLM.
	Content string `json:"content"`
}

// OpenAIClient is an implementation of the Client interface for the OpenAI API.
type OpenAIClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIClient creates a new instance of OpenAIClient.
//
// Parameters:
//   - apiKey: string. The API key for authentication with OpenAI.
//   - baseURL: string. The base URL for the OpenAI API (optional, defaults to "https://api.openai.com/v1").
//
// Returns:
//   - *OpenAIClient: An initialized OpenAIClient.
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

// ChatCompletion performs a chat completion request to the OpenAI API.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: ChatRequest. The chat completion request parameters.
//
// Returns:
//   - *ChatResponse: The chat completion response.
//   - error: An error if the request fails, including network errors or API errors.
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
