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
	t.Run("Default BaseURL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "")
		assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.NotNil(t, client.client)
	})

	t.Run("Custom BaseURL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "https://custom.api.com")
		assert.Equal(t, "https://custom.api.com", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.NotNil(t, client.client)
	})
}

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		req            ChatRequest
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		mockURL        string // Optional: to override server URL (e.g. for network error)
		expectedResp   *ChatResponse
		expectedErr    string
	}{
		{
			name: "Happy Path",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "gpt-4", req.Model)
				assert.Len(t, req.Messages, 1)

				w.WriteHeader(http.StatusOK)
				// Respond with a valid OpenAI response
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
								Content: "Hi there!",
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hi there!",
			},
		},
		{
			name: "API Error (500)",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectedErr: "openai api error (status 500): Internal Server Error",
		},
		{
			name: "JSON Decode Error",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Invalid JSON"))
			},
			expectedErr: "failed to decode response",
		},
		{
			name: "OpenAI Logic Error",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Rate limit exceeded",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "openai error: Rate limit exceeded",
		},
		{
			name: "Empty Choices",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{}, // Empty slice
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "no choices returned",
		},
		{
			name: "Network Error",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {},
			mockURL:     "http://invalid-url.local",
			expectedErr: "request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			baseURL := server.URL
			if tt.mockURL != "" {
				baseURL = tt.mockURL
			}

			client := NewOpenAIClient("test-key", baseURL)
			// Ensure we use the client from the struct which has the timeout set
			// In test environment, we might want to override the http client if we needed to mock RoundTripper,
			// but here httptest.NewServer works fine with default http.Client or the one created in NewOpenAIClient.
			// Note: NewOpenAIClient creates a client with 30s timeout. This is fine for httptest.

			resp, err := client.ChatCompletion(context.Background(), tt.req)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
