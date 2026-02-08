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

func TestNewOpenAIClient(t *testing.T) {
	// Case 1: Default Base URL
	client := NewOpenAIClient("test-key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.Equal(t, "test-key", client.apiKey)

	// Case 2: Custom Base URL
	client = NewOpenAIClient("test-key", "https://custom.api.com")
	assert.Equal(t, "https://custom.api.com", client.baseURL)
	assert.Equal(t, "test-key", client.apiKey)
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		req            ChatRequest
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedResp   *ChatResponse
		expectedErrMsg string
		ctxTimeout     time.Duration
	}{
		{
			name: "Success (Happy Path)",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)

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
			expectedResp: &ChatResponse{
				Content: "Hello there!",
			},
		},
		{
			name: "API Error (Non-200)",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid API Key"))
			},
			expectedErrMsg: "openai api error (status 401): Invalid API Key",
		},
		{
			name: "Invalid JSON Response",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Not JSON"))
			},
			expectedErrMsg: "failed to decode response",
		},
		{
			name: "OpenAI Error Object",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				// We return 200 OK here to specifically test the client's logic that checks
				// for an "error" field in the JSON response body even when the status is 200.
				resp := map[string]interface{}{
					"error": map[string]string{
						"message": "Model not found",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErrMsg: "openai error: Model not found",
		},
		{
			name: "No Choices",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"choices": []interface{}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErrMsg: "no choices returned",
		},
		{
			name: "Context Cancellation",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
			},
			ctxTimeout:     1 * time.Millisecond,
			expectedErrMsg: "request failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.mockHandler))
			defer server.Close()

			// In the Network Error case, we might want to close the server immediately
			// or set an invalid URL. For now, let's handle "Network Error" as a separate test
			// outside this loop if we need specific control, or just rely on Context Cancellation
			// which simulates a form of client-side timeout.

			client := NewOpenAIClient("test-key", server.URL)

			ctx := context.Background()
			if tc.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tc.req)

			if tc.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestChatCompletion_NetworkError(t *testing.T) {
	// Create a client pointing to a closed port/invalid URL
	client := NewOpenAIClient("test-key", "http://127.0.0.1:0")

	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
	assert.Nil(t, resp)
}
