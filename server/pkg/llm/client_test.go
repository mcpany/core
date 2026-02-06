// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		request        ChatRequest
		mockResponse   interface{} // String or struct to be marshaled
		mockStatusCode int
		expectError    bool
		errorContains  string
		expectContent  string
	}{
		{
			name:   "Happy Path",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse: openAIChatResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: "Hello there!",
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectContent:  "Hello there!",
		},
		{
			name:   "API Error - 401 Unauthorized",
			apiKey: "invalid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse:   `{"error": {"message": "Invalid API key"}}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
			errorContains:  "openai api error (status 401)",
		},
		{
			name:   "OpenAI Specific Error",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse: openAIChatResponse{
				Error: &struct {
					Message string `json:"message"`
				}{
					Message: "Rate limit exceeded",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "openai error: Rate limit exceeded",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse: openAIChatResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "no choices returned",
		},
		{
			name:   "Malformed JSON Response",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse:   `{invalid-json`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "failed to decode response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				assert.Equal(t, "Bearer "+tc.apiKey, r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)

				// Verify request body
				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				require.NoError(t, err)
				assert.Equal(t, tc.request.Model, reqBody.Model)
				assert.Equal(t, tc.request.Messages, reqBody.Messages)

				// Send mock response
				w.WriteHeader(tc.mockStatusCode)
				if strResp, ok := tc.mockResponse.(string); ok {
					_, _ = w.Write([]byte(strResp))
				} else {
					err := json.NewEncoder(w).Encode(tc.mockResponse)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			// Initialize client with mock server URL
			client := NewOpenAIClient(tc.apiKey, server.URL)

			// Execute ChatCompletion
			resp, err := client.ChatCompletion(context.Background(), tc.request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expectContent, resp.Content)
			}
		})
	}
}

func TestNewOpenAIClient(t *testing.T) {
	t.Run("Default Base URL", func(t *testing.T) {
		client := NewOpenAIClient("key", "")
		assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
		assert.Equal(t, "key", client.apiKey)
		assert.NotNil(t, client.client)
	})

	t.Run("Custom Base URL", func(t *testing.T) {
		client := NewOpenAIClient("key", "https://custom.api.com")
		assert.Equal(t, "https://custom.api.com", client.baseURL)
		assert.Equal(t, "key", client.apiKey)
		assert.NotNil(t, client.client)
	})
}
