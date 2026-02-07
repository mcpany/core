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

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		req            ChatRequest
		mockHandler    func(t *testing.T, w http.ResponseWriter, r *http.Request)
		expectedResp   *ChatResponse
		expectedErrMsg string
	}{
		{
			name:   "Success",
			apiKey: "test-api-key",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody["model"])

				messages := reqBody["messages"].([]interface{})
				assert.Equal(t, 1, len(messages))
				msg := messages[0].(map[string]interface{})
				assert.Equal(t, "user", msg["role"])
				assert.Equal(t, "Hello", msg["content"])

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
				json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hi there!",
			},
		},
		{
			name:   "OpenAI Error Response",
			apiKey: "test-api-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Invalid API Key",
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			expectedErrMsg: "openai error: Invalid API Key",
		},
		{
			name:   "HTTP Error 401",
			apiKey: "test-api-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
			},
			expectedErrMsg: "openai api error (status 401): Unauthorized",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-api-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"choices": []interface{}{},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			expectedErrMsg: "no choices returned",
		},
		{
			name:   "Invalid JSON Response",
			apiKey: "test-api-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{invalid-json"))
			},
			expectedErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.mockHandler(t, w, r)
			}))
			defer server.Close()

			client := NewOpenAIClient(tt.apiKey, server.URL)

			resp, err := client.ChatCompletion(context.Background(), tt.req)

			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestChatCompletion_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOpenAIClient("key", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.ChatCompletion(ctx, ChatRequest{Model: "gpt-4"})
	require.Error(t, err)
	// Error message for context cancellation can vary ("context deadline exceeded" or "request canceled")
	// assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestNewOpenAIClient_DefaultURL(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.Equal(t, "key", client.apiKey)
	assert.NotNil(t, client.client)
}
