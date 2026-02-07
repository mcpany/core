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
		mockResponse   interface{}
		mockStatusCode int
		mockBody       string // Overrides mockResponse if set
		baseURL        string // Overrides server URL if set
		expectError    bool
		errorContains  string
		expected       *ChatResponse
	}{
		{
			name: "Happy Path",
			mockResponse: openAIChatResponse{
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
			},
			mockStatusCode: http.StatusOK,
			expected: &ChatResponse{
				Content: "Hello, world!",
			},
		},
		{
			name:           "API Error (Non-200)",
			mockBody:       `{"error": {"message": "Invalid API Key"}}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
			errorContains:  "openai api error (status 401)",
		},
		{
			name:           "Malformed Response",
			mockBody:       `{invalid json}`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "failed to decode response",
		},
		{
			name: "API Logic Error",
			mockResponse: openAIChatResponse{
				Error: &struct {
					Message string `json:"message"`
				}{
					Message: "Rate limit exceeded",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "openai error: Rate limit exceeded",
		},
		{
			name: "Empty Choices",
			mockResponse: openAIChatResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorContains:  "no choices returned",
		},
		{
			name:          "Network Error",
			baseURL:       "http://invalid-url:1234",
			expectError:   true,
			errorContains: "request failed",
		},
		{
			name:          "Request Creation Error",
			baseURL:       "http://\x00", // Invalid control character in URL
			expectError:   true,
			errorContains: "failed to create request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
				assert.Equal(t, "/chat/completions", r.URL.Path)

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockBody != "" {
					_, _ = w.Write([]byte(tt.mockBody))
				} else {
					_ = json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			baseURL := server.URL
			if tt.baseURL != "" {
				baseURL = tt.baseURL
			}

			client := NewOpenAIClient("test-api-key", baseURL)
			// Lower timeout for network error test speed
			if tt.name == "Network Error" {
				client.client.Timeout = 100 * time.Millisecond
			}

			req := ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			}

			resp, err := client.ChatCompletion(context.Background(), req)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
		})
	}
}
