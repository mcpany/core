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
		req            ChatRequest
		mockStatus     int
		mockBody       string
		mockDelay      time.Duration
		expectError    bool
		errorContains  string
		expectedOutput string
		validateReq    func(*testing.T, *http.Request)
	}{
		{
			name: "Happy Path",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"choices": [
					{
						"message": {
							"content": "Hi there!"
						}
					}
				]
			}`,
			expectedOutput: "Hi there!",
			validateReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

				var body openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, "gpt-4", body.Model)
				assert.Len(t, body.Messages, 1)
				assert.Equal(t, "Hello", body.Messages[0].Content)
			},
		},
		{
			name: "API Error (500)",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus:    http.StatusInternalServerError,
			mockBody:      `Internal Server Error`,
			expectError:   true,
			errorContains: "openai api error (status 500)",
		},
		{
			name: "Structured Error in Response",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus: http.StatusOK, // Some APIs return 200 with error field, or we might mock 400
			mockBody: `{
				"error": {
					"message": "Invalid model"
				}
			}`,
			expectError:   true,
			errorContains: "openai error: Invalid model",
		},
		{
			name: "Empty Choices",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"choices": []
			}`,
			expectError:   true,
			errorContains: "no choices returned",
		},
		{
			name: "Invalid JSON Response",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus:    http.StatusOK,
			mockBody:      `{ invalid json `,
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name: "Context Cancellation",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus:  http.StatusOK,
			mockDelay:   100 * time.Millisecond, // Delay longer than context timeout
			expectError: true,
			// Error message depends on how http client handles it, usually "context deadline exceeded" or "canceled"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateReq != nil {
					tt.validateReq(t, r)
				}
				if tt.mockDelay > 0 {
					time.Sleep(tt.mockDelay)
				}
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.name == "Context Cancellation" {
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, resp.Content)
			}
		})
	}
}
