// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		req            llm.ChatRequest
		mockStatus     int
		mockResponse   interface{} // String or struct
		mockDelay      time.Duration
		ctxTimeout     time.Duration
		expectedResult *llm.ChatResponse
		expectedError  string
	}{
		{
			name:   "Success",
			apiKey: "test-key",
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"choices": []interface{}{
					map[string]interface{}{
						"message": map[string]interface{}{
							"content": "Hi there!",
						},
					},
				},
			},
			expectedResult: &llm.ChatResponse{
				Content: "Hi there!",
			},
		},
		{
			name:   "OpenAI Error Response (400)",
			apiKey: "test-key",
			req: llm.ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusBadRequest,
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid model",
				},
			},
			expectedError: "openai api error (status 400)",
		},
		{
			name:   "OpenAI Error Response (200)",
			apiKey: "test-key",
			req: llm.ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Invalid model",
				},
			},
			expectedError: "openai error: Invalid model",
		},
		{
			name:   "Non-200 Status with Body",
			apiKey: "test-key",
			req: llm.ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusInternalServerError,
			mockResponse: map[string]interface{}{
				"some": "error",
			},
			expectedError: "openai api error (status 500)",
		},
		{
			name:   "Invalid JSON Response",
			apiKey: "test-key",
			req: llm.ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockResponse: "{invalid-json", // String to force raw write
			expectedError: "failed to decode response",
		},
		{
			name:   "Empty Choices",
			apiKey: "test-key",
			req: llm.ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"choices": []interface{}{},
			},
			expectedError: "no choices returned",
		},
		{
			name:       "Context Timeout",
			apiKey:     "test-key",
			req:        llm.ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"choices": []interface{}{
					map[string]interface{}{
						"message": map[string]interface{}{
							"content": "Hi",
						},
					},
				},
			},
			mockDelay:  100 * time.Millisecond,
			ctxTimeout: 10 * time.Millisecond,
			expectedError: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify Headers
				assert.Equal(t, "Bearer "+tt.apiKey, r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "/chat/completions", r.URL.Path)

				// Verify Body
				var body llm.ChatRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.req, body)

				if tt.mockDelay > 0 {
					time.Sleep(tt.mockDelay)
				}

				w.WriteHeader(tt.mockStatus)
				if str, ok := tt.mockResponse.(string); ok {
					w.Write([]byte(str))
				} else {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := llm.NewOpenAIClient(tt.apiKey, server.URL)

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
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, resp)
			}
		})
	}
}

func TestNewOpenAIClient_DefaultURL(t *testing.T) {
	client := llm.NewOpenAIClient("key", "")
	// We can't easily check the private field baseURL without reflection or exposing it.
	// But we can check that it doesn't panic.
	// To verify the default URL, we'd need to mock the transport and check the request URL.
	// For now, this just covers the function call.
	require.NotNil(t, client)
}
