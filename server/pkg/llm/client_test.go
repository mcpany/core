// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIClient(t *testing.T) {
	t.Run("default_base_url", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "")
		require.NotNil(t, client)
		// Accessing private fields for verification (in same package)
		assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.NotNil(t, client.client)
	})

	t.Run("custom_base_url", func(t *testing.T) {
		customURL := "https://custom.openai.com/v1"
		client := NewOpenAIClient("test-key", customURL)
		require.NotNil(t, client)
		assert.Equal(t, customURL, client.baseURL)
	})
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		request        ChatRequest
		mockResponse   string
		mockStatusCode int
		expectError    bool
		errorContains  string
		expectContent  string
	}{
		{
			name:   "Happy Path",
			apiKey: "valid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"choices": [{"message": {"content": "Hello world"}}]}`,
			expectError:    false,
			expectContent:  "Hello world",
		},
		{
			name:   "API Error 401",
			apiKey: "invalid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatusCode: http.StatusUnauthorized,
			mockResponse:   `{"error": {"message": "Invalid API Key"}}`,
			expectError:    true,
			errorContains:  "openai api error (status 401)",
		},
		{
			name:   "API Error 500",
			apiKey: "valid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   "Internal Server Error",
			expectError:    true,
			errorContains:  "openai api error (status 500)",
		},
		{
			name:   "Logic Error (Error field present)",
			apiKey: "valid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"error": {"message": "Rate limit exceeded"}}`,
			expectError:    true,
			errorContains:  "openai error: Rate limit exceeded",
		},
		{
			name:   "Invalid JSON",
			apiKey: "valid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `invalid-json`,
			expectError:    true,
			errorContains:  "failed to decode response",
		},
		{
			name:   "No Choices",
			apiKey: "valid-key",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"choices": []}`,
			expectError:    true,
			errorContains:  "no choices returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify Method
				assert.Equal(t, "POST", r.Method)
				// Verify URL
				assert.Equal(t, "/chat/completions", r.URL.Path)
				// Verify Headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer "+tt.apiKey, r.Header.Get("Authorization"))

				// Decode Request Body to verify it matches input
				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.request.Model, reqBody.Model)
				assert.Equal(t, len(tt.request.Messages), len(reqBody.Messages))

				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := NewOpenAIClient(tt.apiKey, server.URL)
			// Override client timeout for faster tests if needed, though mock server is local.
			client.client.Timeout = 2 * time.Second

			resp, err := client.ChatCompletion(context.Background(), tt.request)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectContent, resp.Content)
			}
		})
	}
}
