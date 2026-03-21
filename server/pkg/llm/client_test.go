// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIClient(t *testing.T) {
	client := NewOpenAIClient("test-key", "")
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.NotNil(t, client.client)
	assert.Equal(t, 30*time.Second, client.client.Timeout)

	client2 := NewOpenAIClient("test-key-2", "http://custom-url")
	assert.Equal(t, "test-key-2", client2.apiKey)
	assert.Equal(t, "http://custom-url", client2.baseURL)
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		req            ChatRequest
		mockStatus     int
		mockResponse   string
		mockDelay      time.Duration
		cancelContext  bool
		expectedResult *ChatResponse
		expectedError  string
		validateReq    func(t *testing.T, r *http.Request)
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
			mockResponse: `{
				"choices": [
					{
						"message": {
							"content": "Hi there!"
						}
					}
				]
			}`,
			expectedResult: &ChatResponse{
				Content: "Hi there!",
			},
			validateReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var reqBody map[string]interface{}
				err = json.Unmarshal(body, &reqBody)
				require.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody["model"])
			},
		},
		{
			name: "OpenAI Error Response",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus: http.StatusOK, // Some APIs return 200 even with logical errors in body, but let's stick to what the code handles. The code handles 200 OK and then checks for `error` field.
			mockResponse: `{
				"error": {
					"message": "Invalid API Key"
				}
			}`,
			expectedError: "openai error: Invalid API Key",
		},
		{
			name: "HTTP Error Status",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus:    http.StatusUnauthorized,
			mockResponse:  `Unauthorized`,
			expectedError: "openai api error (status 401): Unauthorized",
		},
		{
			name: "Invalid JSON Response",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus:    http.StatusOK,
			mockResponse:  `{invalid-json`,
			expectedError: "failed to decode response",
		},
		{
			name: "Empty Choices",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus: http.StatusOK,
			mockResponse: `{
				"choices": []
			}`,
			expectedError: "no choices returned",
		},
		{
			name: "Context Cancellation",
			req: ChatRequest{
				Model: "gpt-4",
			},
			mockStatus:    http.StatusOK,
			mockDelay:     500 * time.Millisecond,
			cancelContext: true,
			expectedError: "request failed",
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
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			ctx := context.Background()
			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, resp)
			}
		})
	}
}
