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
		handler        func(w http.ResponseWriter, r *http.Request)
		ctxCancel      bool
		expectedResp   *ChatResponse
		expectedErr    string
	}{
		{
			name:   "Happy Path",
			apiKey: "test-api-key",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", body.Model)
				assert.Len(t, body.Messages, 1)

				w.WriteHeader(http.StatusOK)

				resp := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"content": "Hello there!",
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hello there!",
			},
		},
		{
			name:   "OpenAI API Error",
			apiKey: "test-api-key",
			req:    ChatRequest{Model: "gpt-4"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API Key"}}`))
			},
			expectedErr: "openai api error (status 401): {\"error\": {\"message\": \"Invalid API Key\"}}",
		},
		{
			name:   "Malformed JSON Response",
			apiKey: "test-api-key",
			req:    ChatRequest{Model: "gpt-4"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid-json`))
			},
			expectedErr: "failed to decode response",
		},
		{
			name:   "OpenAI Logical Error",
			apiKey: "test-api-key",
			req:    ChatRequest{Model: "gpt-4"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"error": {"message": "Rate limit exceeded"}}`))
			},
			expectedErr: "openai error: Rate limit exceeded",
		},
		{
			name:   "No Choices Returned",
			apiKey: "test-api-key",
			req:    ChatRequest{Model: "gpt-4"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"choices": []}`))
			},
			expectedErr: "no choices returned",
		},
		{
			name:      "Context Cancelled",
			apiKey:    "test-api-key",
			req:       ChatRequest{Model: "gpt-4"},
			ctxCancel: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond) // Ensure context cancels before we respond
				w.WriteHeader(http.StatusOK)
			},
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			client := NewOpenAIClient(tt.apiKey, server.URL)

			ctx := context.Background()
			if tt.ctxCancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
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

func TestNewOpenAIClient_DefaultURL(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
}
