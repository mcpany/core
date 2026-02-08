// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewOpenAIClient tests the initialization of the OpenAI client.
func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		baseURL  string
		expected string
	}{
		{
			name:     "Default URL",
			apiKey:   "test-key",
			baseURL:  "",
			expected: "https://api.openai.com/v1",
		},
		{
			name:     "Custom URL",
			apiKey:   "test-key",
			baseURL:  "http://custom-url.com",
			expected: "http://custom-url.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL)
			if client.baseURL != tt.expected {
				t.Errorf("expected baseURL %q, got %q", tt.expected, client.baseURL)
			}
			if client.apiKey != tt.apiKey {
				t.Errorf("expected apiKey %q, got %q", tt.apiKey, client.apiKey)
			}
		})
	}
}

// TestChatCompletion tests the ChatCompletion method.
func TestChatCompletion(t *testing.T) {
	type testCase struct {
		name            string
		mockHandler     func(w http.ResponseWriter, r *http.Request)
		req             ChatRequest
		expectedContent string
		expectError     bool
		errorContains   string
		ctxTimeout      time.Duration
	}

	tests := []testCase{
		{
			name: "Success",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				if r.URL.Path != "/chat/completions" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				resp := map[string]interface{}{
					"choices": []map[string]interface{}{
						{
							"message": map[string]interface{}{
								"content": "Hello",
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hi"},
				},
			},
			expectedContent: "Hello",
			expectError:     false,
		},
		{
			name: "API Error 500",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			req:           ChatRequest{},
			expectError:   true,
			errorContains: "openai api error (status 500)",
		},
		{
			name: "OpenAI Error Response",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Invalid API Key",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			req:           ChatRequest{},
			expectError:   true,
			errorContains: "openai error: Invalid API Key",
		},
		{
			name: "Empty Choices",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"choices": []interface{}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			req:           ChatRequest{},
			expectError:   true,
			errorContains: "no choices returned",
		},
		{
			name: "Invalid JSON",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("invalid-json"))
			},
			req:           ChatRequest{},
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name: "Context Cancelled",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
			},
			req:           ChatRequest{},
			expectError:   true,
			errorContains: "context deadline exceeded",
			ctxTimeout:    50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.Content != tt.expectedContent {
					t.Errorf("expected content %q, got %q", tt.expectedContent, resp.Content)
				}
			}
		})
	}
}

// TestChatCompletion_NetworkError tests the client behavior when the server is unreachable.
func TestChatCompletion_NetworkError(t *testing.T) {
	// Create a server and close it immediately to simulate a connection error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	client := NewOpenAIClient("test-key", url)
	_, err := client.ChatCompletion(context.Background(), ChatRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
