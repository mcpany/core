// Copyright 2025 Author(s) of MCP Any
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
		mockStatus     int
		mockResponse   string // JSON response body
		mockDelay      time.Duration
		ctxTimeout     time.Duration
		req            llm.ChatRequest
		wantErr        bool
		wantErrMessage string
		wantContent    string
	}{
		{
			name:       "Happy Path",
			mockStatus: http.StatusOK,
			mockResponse: `{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "Hello, world!"
						}
					}
				]
			}`,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:     false,
			wantContent: "Hello, world!",
		},
		{
			name:       "OpenAI Error Response",
			mockStatus: http.StatusOK,
			mockResponse: `{
				"error": {
					"message": "Invalid API Key"
				}
			}`,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:        true,
			wantErrMessage: "openai error: Invalid API Key",
		},
		{
			name:       "Empty Choices",
			mockStatus: http.StatusOK,
			mockResponse: `{
				"choices": []
			}`,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:        true,
			wantErrMessage: "no choices returned",
		},
		{
			name:       "Invalid JSON Response",
			mockStatus: http.StatusOK,
			mockResponse: `{
				"choices": [
					{ "message": ... invalid json ... }
			}`,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:        true,
			wantErrMessage: "failed to decode response",
		},
		{
			name:       "HTTP 401 Unauthorized",
			mockStatus: http.StatusUnauthorized,
			mockResponse: `{
				"error": {
					"message": "Incorrect API key provided"
				}
			}`,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:        true,
			wantErrMessage: "openai api error (status 401):",
		},
		{
			name:       "HTTP 500 Internal Server Error",
			mockStatus: http.StatusInternalServerError,
			mockResponse: `Internal Server Error`,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:        true,
			wantErrMessage: "openai api error (status 500): Internal Server Error",
		},
		{
			name:       "Context Deadline Exceeded",
			mockStatus: http.StatusOK,
			mockDelay:  100 * time.Millisecond,
			ctxTimeout: 10 * time.Millisecond,
			req: llm.ChatRequest{
				Model: "gpt-4",
				Messages: []llm.Message{
					{Role: "user", Content: "Hi"},
				},
			},
			wantErr:        true,
			wantErrMessage: "context deadline exceeded", // Standard error message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and headers
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

				// Verify request body
				var reqBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.req.Model, reqBody["model"])

				// Simulate delay
				if tt.mockDelay > 0 {
					time.Sleep(tt.mockDelay)
				}

				// Respond
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create client with mock server URL
			client := llm.NewOpenAIClient("test-api-key", server.URL)

			// Create context
			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			// Call method
			got, err := client.ChatCompletion(ctx, tt.req)

			// Assert results
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMessage != "" {
					assert.Contains(t, err.Error(), tt.wantErrMessage)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantContent, got.Content)
			}
		})
	}
}
