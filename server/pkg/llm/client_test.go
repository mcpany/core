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
		req            ChatRequest
		handler        http.HandlerFunc
		expectedResp   *ChatResponse
		expectedErr    string
		expectError    bool
		timeoutContext bool
	}{
		{
			name:   "Happy Path",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", req.Model)
				assert.Equal(t, "user", req.Messages[0].Role)
				assert.Equal(t, "Hello", req.Messages[0].Content)

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
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hi there!",
			},
			expectError: false,
		},
		{
			name:   "Non-200 Status Code",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API Key"}}`))
			},
			expectedErr: "openai api error (status 401)",
			expectError: true,
		},
		{
			name:   "Invalid JSON Response",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid-json`))
			},
			expectedErr: "failed to decode response",
			expectError: true,
		},
		{
			name:   "OpenAI Error Response",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
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
			expectedErr: "openai error: Rate limit exceeded",
			expectError: true,
		},
		{
			name:   "Empty Choices Response",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
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
			expectedErr: "no choices returned",
			expectError: true,
		},
		{
			name:   "Context Cancelled",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond) // Wait longer than context timeout
				w.WriteHeader(http.StatusOK)
			},
			expectError:    true,
			timeoutContext: true,
			expectedErr:    "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewOpenAIClient(tt.apiKey, server.URL)

			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.timeoutContext {
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
