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
		request        ChatRequest
		mockHandler    http.HandlerFunc
		expectedResp   *ChatResponse
		expectedErr    string
		expectRequest  bool
		validateRequest func(*testing.T, *http.Request)
	}{
		{
			name: "Happy Path",
			request: ChatRequest{
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

				var body openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, "gpt-4", body.Model)
				assert.Len(t, body.Messages, 1)
				assert.Equal(t, "Hello", body.Messages[0].Content)

				resp := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"content": "Hi there!",
							},
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hi there!",
			},
			expectRequest: true,
		},
		{
			name: "API Error (500)",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
			},
			expectedErr:   "openai api error (status 500): Internal Server Error",
			expectRequest: true,
		},
		{
			name: "API Error (200 with Error field)",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Invalid API Key",
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedErr:   "openai error: Invalid API Key",
			expectRequest: true,
		},
		{
			name: "Malformed JSON Response",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{invalid-json"))
			},
			expectedErr:   "failed to decode response",
			expectRequest: true,
		},
		{
			name: "Empty Choices",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"choices": []interface{}{},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedErr:   "no choices returned",
			expectRequest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.mockHandler)
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.ChatCompletion(ctx, tt.request)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestOpenAIClient_NetworkError(t *testing.T) {
	// Use a closed port or invalid URL to simulate connection error
	client := NewOpenAIClient("test-key", "http://127.0.0.1:0")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.ChatCompletion(ctx, ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}
