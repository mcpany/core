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
		handler        func(w http.ResponseWriter, r *http.Request)
		ctxTimeout     time.Duration
		expectedResp   *ChatResponse
		expectedErr    string
		expectReqCheck func(*testing.T, *http.Request)
	}{
		{
			name:   "Success",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(openAIChatResponse{
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
				})
			},
			expectReqCheck: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, "gpt-4", body.Model)
				assert.Len(t, body.Messages, 1)
				assert.Equal(t, "Hello", body.Messages[0].Content)
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
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API Key"}}`))
			},
			expectedErr: "openai api error (status 401): {\"error\": {\"message\": \"Invalid API Key\"}}",
		},
		{
			name:   "API Logical Error (200 OK but error field)",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				// Some APIs might return 200 but include an error object (though OpenAI usually uses status codes)
				// The client code explicitly checks for `response.Error`
				json.NewEncoder(w).Encode(openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Something went wrong",
					},
				})
			},
			expectedErr: "openai error: Something went wrong",
		},
		{
			name:   "Malformed JSON Response",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid-json`))
			},
			expectedErr: "failed to decode response",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-key",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{}, // Empty slice
				})
			},
			expectedErr: "no choices returned",
		},
		{
			name: "Network Error (Context Cancelled)",
			req: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond) // Wait longer than timeout
			},
			ctxTimeout:  50 * time.Millisecond,
			expectedErr: "request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectReqCheck != nil {
					tt.expectReqCheck(t, r)
				}
				tt.handler(w, r)
			}))
			defer ts.Close()

			// Initialize client with mock server URL
			client := NewOpenAIClient(tt.apiKey, ts.URL)

			// Setup context
			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			// Execute
			resp, err := client.ChatCompletion(ctx, tt.req)

			// Verify
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
