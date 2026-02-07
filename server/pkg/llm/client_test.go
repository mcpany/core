// Copyright 2026 Author(s) of MCP Any
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
	t.Parallel()

	t.Run("Default URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "")
		assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.Equal(t, 30*time.Second, client.client.Timeout)
	})

	t.Run("Custom URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "https://custom.api.com")
		assert.Equal(t, "https://custom.api.com", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
	})
}

func TestChatCompletion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		req          ChatRequest
		mockHandler  func(w http.ResponseWriter, r *http.Request)
		expectedResp *ChatResponse
		expectError  bool
		errorContains string
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
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)
				assert.Len(t, reqBody.Messages, 1)

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
			expectError: false,
		},
		{
			name: "API Error (400)",
			req: ChatRequest{Model: "gpt-4"},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": {"message": "Invalid request"}}`))
			},
			expectError: true,
			errorContains: "openai api error (status 400)",
		},
		{
			name: "OpenAI Error (200 OK but error field)",
			req: ChatRequest{Model: "gpt-4"},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Some logical error",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectError: true,
			errorContains: "openai error: Some logical error",
		},
		{
			name: "Invalid JSON Response",
			req: ChatRequest{Model: "gpt-4"},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{invalid-json`))
			},
			expectError: true,
			errorContains: "failed to decode response",
		},
		{
			name: "Empty Choices",
			req: ChatRequest{Model: "gpt-4"},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectError: true,
			errorContains: "no choices returned",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			ctx := context.Background()
			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}

	t.Run("Network Error", func(t *testing.T) {
		t.Parallel()
		// Create a client with an invalid URL to simulate network error
		client := NewOpenAIClient("test-key", "http://localhost:0") // port 0 is invalid usually or closed
		// Actually "http://invalid-url" might be better but relies on DNS.
		// A closed port on localhost is safer.
		// But Wait, "http://localhost:0" might be resolved but connection refused.
		// Let's use a very short timeout and a non-routable IP to force failure?
		// Or just close the server immediately?
		// If I don't start a server, and point to a closed port.

		client = NewOpenAIClient("test-key", "http://127.0.0.1:1") // Port 1 is likely closed
		client.client.Timeout = 100 * time.Millisecond // fast fail

		_, err := client.ChatCompletion(context.Background(), ChatRequest{Model: "gpt-4"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Wait longer than the context
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.ChatCompletion(ctx, ChatRequest{Model: "gpt-4"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}
