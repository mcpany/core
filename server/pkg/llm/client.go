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
// It defines the standard operations available for interacting with Large Language Model providers.
type Client interface {
	// ChatCompletion performs a chat completion request.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - req: ChatRequest. The request parameters.
	//
	// Returns:
	//   - *ChatResponse: The completion response.
	//   - error: An error if the operation fails.
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ChatRequest represents a chat completion request to an LLM provider.
//
// It encapsulates the model selection and the history of messages for the conversation context.
type ChatRequest struct {
	// Model is the identifier of the LLM model to use (e.g., "gpt-4o", "claude-3-5-sonnet").
	Model string `json:"model"`
	// Messages is the list of messages in the conversation history, including the latest user message.
	Messages []Message `json:"messages"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	// Role indicates the sender of the message (e.g., "user", "assistant", "system").
	Role string `json:"role"`
	// Content is the text content of the message.
	Content string `json:"content"`
}

// ChatResponse represents the response returned by a chat completion request.
type ChatResponse struct {
	// Content is the generated text response from the LLM.
	Content string `json:"content"`
}

// OpenAIClient implements the Client interface for the OpenAI API.
//
// It handles authentication and communication with OpenAI-compatible endpoints.
type OpenAIClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIClient initializes a new client for interacting with the OpenAI API or compatible endpoints.
//
// Parameters:
//   - apiKey: string. The API key for authentication.
//   - baseURL: string. The base URL of the API. If empty, defaults to "https://api.openai.com/v1".
//
// Returns:
//   - *OpenAIClient: A pointer to the initialized OpenAIClient.
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

// ChatCompletion sends a chat completion request to the OpenAI API and returns the generated response.
//
// Parameters:
//   - ctx: context.Context. The context for the request, allowing for cancellation and timeouts.
//   - req: ChatRequest. The request payload containing the model and messages.
//
// Returns:
//   - *ChatResponse: The generated response from the LLM.
//   - error: An error if the request fails, including network errors, API errors, or parsing issues.
//
// Errors:
//   - Returns error if request marshaling fails.
//   - Returns error if creating the HTTP request fails.
//   - Returns error if the network request fails or times out.
//   - Returns error if the API returns a non-200 status code.
//   - Returns error if the response body cannot be decoded.
//   - Returns error if the API response contains an error message.
//
// Side Effects:
//   - Makes an HTTP POST request to the configured base URL.
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
