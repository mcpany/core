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

func TestNewOpenAIClient(t *testing.T) {
	t.Run("default base URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "")
		assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.NotNil(t, client.client)
	})

	t.Run("custom base URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "https://custom.openai.com")
		assert.Equal(t, "https://custom.openai.com", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.NotNil(t, client.client)
	})
}

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   func(t *testing.T, w http.ResponseWriter, r *http.Request)
		input          ChatRequest
		expected       *ChatResponse
		expectedErr    string
		expectError    bool
	}{
		{
			name: "success",
			mockResponse: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", req.Model)
				assert.Len(t, req.Messages, 1)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{
						{
							Message: struct {
								Content string `json:"content"`
							}{
								Content: "Hello!",
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			input: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hi"},
				},
			},
			expected: &ChatResponse{
				Content: "Hello!",
			},
			expectError: false,
		},
		{
			name: "api error 401",
			mockResponse: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
			},
			input: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{{Role: "user", Content: "Hi"}},
			},
			expectError: true,
			expectedErr: "openai api error (status 401)",
		},
		{
			name: "api error with error field in 200 OK (unlikely but guarded)",
			mockResponse: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Something went wrong",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			input: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{{Role: "user", Content: "Hi"}},
			},
			expectError: true,
			expectedErr: "openai error: Something went wrong",
		},
		{
			name: "empty choices",
			mockResponse: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			input: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{{Role: "user", Content: "Hi"}},
			},
			expectError: true,
			expectedErr: "no choices returned",
		},
		{
			name: "malformed json response",
			mockResponse: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid-json`))
			},
			input: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{{Role: "user", Content: "Hi"}},
			},
			expectError: true,
			expectedErr: "failed to decode response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.mockResponse(t, w, r)
			}))
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)
			resp, err := client.ChatCompletion(context.Background(), tc.input)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != "" {
					assert.Contains(t, err.Error(), tc.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, resp)
			}
		})
	}
}

func TestOpenAIClient_ChatCompletion_NetworkError(t *testing.T) {
	// Create a client pointing to a closed port to force a connection error
	client := NewOpenAIClient("test-key", "http://127.0.0.1:0")

	resp, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
	assert.Nil(t, resp)
}
