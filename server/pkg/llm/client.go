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

// Client represents the public Client entity.
//
// Summary: Defines the structured data model representing a .
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

// ChatRequest represents the public ChatRequest entity.
//
// Summary: Defines the structured data model representing a request.
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
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents the public Message entity.
//
// Summary: Defines the structured data model representing a .
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
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the public ChatResponse entity.
//
// Summary: Defines the structured data model representing a response.
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
type ChatResponse struct {
	Content string `json:"content"`
}

// OpenAIClient represents the public OpenAIClient entity.
//
// Summary: Defines the structured data model representing a ai client.
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
type OpenAIClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIClient serves as a public interface for interacting with NewOpenAIClient.
//
// Summary: Constructs and returns an initialized open ai client ready for consumption.
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

// ChatCompletion serves as a public interface for interacting with ChatCompletion.
//
// Summary: Chat the completion appropriately based on current system conditions.
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
