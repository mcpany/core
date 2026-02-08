// Copyright 2025 Author(s) of MCP Any
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
	tests := []struct {
		name            string
		apiKey          string
		baseURL         string
		expectedBaseURL string
	}{
		{
			name:            "default base url",
			apiKey:          "test-key",
			baseURL:         "",
			expectedBaseURL: "https://api.openai.com/v1",
		},
		{
			name:            "custom base url",
			apiKey:          "test-key",
			baseURL:         "https://custom.openai.com/v1",
			expectedBaseURL: "https://custom.openai.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL)
			assert.Equal(t, tt.apiKey, client.apiKey)
			assert.Equal(t, tt.expectedBaseURL, client.baseURL)
			assert.NotNil(t, client.client)
		})
	}
}

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	tests := []struct {
		name             string
		mockStatus       int
		mockBody         string
		mockDelay        time.Duration
		contextTimeout   time.Duration
		expectError      bool
		errorContains    string
		expectedResponse *ChatResponse
		validateRequest  func(t *testing.T, r *http.Request)
	}{
		{
			name:       "success",
			mockStatus: http.StatusOK,
			mockBody: `{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "Hello, world!"
						}
					}
				]
			}`,
			expectedResponse: &ChatResponse{
				Content: "Hello, world!",
			},
			validateRequest: func(t *testing.T, r *http.Request) {
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
				messages, ok := reqBody["messages"].([]interface{})
				require.True(t, ok)
				require.Len(t, messages, 1)

				msg, ok := messages[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "user", msg["role"])
				assert.Equal(t, "Hello", msg["content"])
			},
		},
		{
			name:          "api error 401",
			mockStatus:    http.StatusUnauthorized,
			mockBody:      `{"error": {"message": "Invalid API Key"}}`,
			expectError:   true,
			errorContains: "openai api error (status 401)",
		},
		{
			name:          "malformed json response",
			mockStatus:    http.StatusOK,
			mockBody:      `{invalid-json}`,
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name:       "openai error field",
			mockStatus: http.StatusOK,
			mockBody: `{
				"error": {
					"message": "Model not found"
				}
			}`,
			expectError:   true,
			errorContains: "openai error: Model not found",
		},
		{
			name:       "no choices returned",
			mockStatus: http.StatusOK,
			mockBody: `{
				"choices": []
			}`,
			expectError:   true,
			errorContains: "no choices returned",
		},
		{
			name:           "context timeout",
			mockStatus:     http.StatusOK,
			mockBody:       `{}`,
			mockDelay:      100 * time.Millisecond,
			contextTimeout: 50 * time.Millisecond,
			expectError:    true,
			errorContains:  "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateRequest != nil {
					tt.validateRequest(t, r)
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
			if tt.contextTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.contextTimeout)
				defer cancel()
			}

			req := ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			}

			resp, err := client.ChatCompletion(ctx, req)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, resp)
			}
		})
	}
}
