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

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		handler        func(w http.ResponseWriter, r *http.Request)
		customBaseURL  string // Optional: if set, overrides the mock server URL
		req            ChatRequest
		ctxTimeout     time.Duration
		expectedResult *ChatResponse
		expectedError  string
	}{
		{
			name:   "Happy Path",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)

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
								Content: "Hello, world!",
							},
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hi"},
				},
			},
			expectedResult: &ChatResponse{
				Content: "Hello, world!",
			},
		},
		{
			name:   "API Error - 401 Unauthorized",
			apiKey: "bad-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
			},
			req: ChatRequest{Model: "gpt-4"},
			expectedError: "openai api error (status 401)",
		},
		{
			name:   "OpenAI Logic Error - 200 OK with Error field",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Rate limit exceeded",
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{Model: "gpt-4"},
			expectedError: "openai error: Rate limit exceeded",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{Model: "gpt-4"},
			expectedError: "no choices returned",
		},
		{
			name:   "Malformed JSON Response",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid-json}`))
			},
			req: ChatRequest{Model: "gpt-4"},
			expectedError: "failed to decode response",
		},
		{
			name:   "Context Timeout",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			req:        ChatRequest{Model: "gpt-4"},
			ctxTimeout: 10 * time.Millisecond,
			expectedError: "request failed", // client.Do returns error on context cancel
		},
		{
			name: "Connection Refused",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {},
			customBaseURL: "http://127.0.0.1:12345", // Assuming this port is closed/unused
			req: ChatRequest{Model: "gpt-4"},
			expectedError: "request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			baseURL := server.URL
			if tt.customBaseURL != "" {
				baseURL = tt.customBaseURL
			}
			client := NewOpenAIClient(tt.apiKey, baseURL)

			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, resp)
			}
		})
	}
}

func TestNewOpenAIClient(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.Equal(t, "key", client.apiKey)

	client = NewOpenAIClient("key", "custom-url")
	assert.Equal(t, "custom-url", client.baseURL)
}
