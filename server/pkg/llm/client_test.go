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
		req            ChatRequest
		handler        http.HandlerFunc
		expectedResp   *ChatResponse
		expectedErr    string
		expectTimeout  bool
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
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				require.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)
				assert.Equal(t, 1, len(reqBody.Messages))
				assert.Equal(t, "Hello", reqBody.Messages[0].Content)

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
			name:   "API Error (401)",
			apiKey: "invalid-key",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
			},
			expectedErr: "openai api error (status 401)",
		},
		{
			name: "OpenAI Logic Error",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Something went wrong",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "openai error: Something went wrong",
		},
		{
			name: "Malformed Response",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{not valid json`))
			},
			expectedErr: "failed to decode response",
		},
		{
			name: "Empty Choices",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "no choices returned",
		},
		{
			name: "Context Timeout",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
			},
			expectTimeout: true,
			expectedErr:   "request failed", // net/http error usually
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewOpenAIClient(tt.apiKey, server.URL)

			ctx := context.Background()
			if tt.expectTimeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

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

func TestNewOpenAIClient_Defaults(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.Equal(t, "key", client.apiKey)

	client2 := NewOpenAIClient("key", "http://custom.url")
	assert.Equal(t, "http://custom.url", client2.baseURL)
}
