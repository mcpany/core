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

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		request        ChatRequest
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedResult *ChatResponse
		expectedError  string
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
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "POST", r.Method)

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)

				w.WriteHeader(http.StatusOK)
				// Response based on openAIChatResponse struct in client.go
				resp := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]string{
								"content": "Hello there!",
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedResult: &ChatResponse{
				Content: "Hello there!",
			},
		},
		{
			name:   "Authentication Error",
			apiKey: "bad-key",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
			},
			expectedError: "openai api error (status 401)",
		},
		{
			name:   "Server Error",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectedError: "openai api error (status 500)",
		},
		{
			name:   "Malformed JSON Response",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid-json`))
			},
			expectedError: "failed to decode response",
		},
		{
			name:   "API Error Response",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := map[string]interface{}{
					"error": map[string]string{
						"message": "Model overload",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedError: "openai error: Model overload",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-key",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				resp := map[string]interface{}{
					"choices": []interface{}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedError: "no choices returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			client := NewOpenAIClient(tt.apiKey, server.URL)
			result, err := client.ChatCompletion(context.Background(), tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestOpenAIClient_ChatCompletion_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate delay
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.ChatCompletion(ctx, ChatRequest{Model: "gpt-4"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestNewOpenAIClient_Defaults(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
}
