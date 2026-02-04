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
)

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		baseURL  string
		expectedBaseURL string
	}{
		{
			name:     "default base URL",
			apiKey:   "test-key",
			baseURL:  "",
			expectedBaseURL: "https://api.openai.com/v1",
		},
		{
			name:     "custom base URL",
			apiKey:   "test-key",
			baseURL:  "https://custom-api.com",
			expectedBaseURL: "https://custom-api.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL)
			if client.apiKey != tt.apiKey {
				t.Errorf("expected apiKey %s, got %s", tt.apiKey, client.apiKey)
			}
			if client.baseURL != tt.expectedBaseURL {
				t.Errorf("expected baseURL %s, got %s", tt.expectedBaseURL, client.baseURL)
			}
			if client.client == nil {
				t.Error("expected http client to be initialized")
			}
		})
	}
}

func TestChatCompletion(t *testing.T) {
	// Helper to create a test server
	createTestServer := func(handler http.HandlerFunc) *httptest.Server {
		return httptest.NewServer(handler)
	}

	tests := []struct {
		name          string
		handler       http.HandlerFunc
		req           ChatRequest
		wantErr       bool
		expectedContent string
		errorContains string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify Method
				if r.Method != http.MethodPost {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				// Verify URL
				if r.URL.Path != "/chat/completions" {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				// Verify Headers
				if r.Header.Get("Content-Type") != "application/json" {
					http.Error(w, "bad content type", http.StatusBadRequest)
					return
				}
				if r.Header.Get("Authorization") != "Bearer test-key" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				// Verify Body
				var req openAIChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if req.Model != "gpt-4" {
					http.Error(w, "wrong model", http.StatusBadRequest)
					return
				}

				resp := openAIChatResponse{
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
				}
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr:         false,
			expectedContent: "Hello, world!",
		},
		{
			name: "api error 401",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
			},
			req: ChatRequest{Model: "gpt-4"},
			wantErr:       true,
			errorContains: "openai api error (status 401)",
		},
		{
			name: "api error with explicit error message",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Something went wrong",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{Model: "gpt-4"},
			wantErr:       true,
			errorContains: "openai error: Something went wrong",
		},
		{
			name: "malformed json response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{not valid json`))
			},
			req: ChatRequest{Model: "gpt-4"},
			wantErr:       true,
			errorContains: "failed to decode response",
		},
		{
			name: "empty choices",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{Model: "gpt-4"},
			wantErr:       true,
			errorContains: "no choices returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(tt.handler)
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			resp, err := client.ChatCompletion(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ChatCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != nil && tt.errorContains != "" {
					if !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("ChatCompletion() error = %v, expected to contain %v", err, tt.errorContains)
					}
				}
				return
			}

			if resp == nil {
				t.Fatal("ChatCompletion() response is nil")
			}
			if resp.Content != tt.expectedContent {
				t.Errorf("ChatCompletion() content = %v, want %v", resp.Content, tt.expectedContent)
			}
		})
	}
}

// Separate test for Network Error to handle server closing
func TestChatCompletion_NetworkError(t *testing.T) {
	// Create a server that closes immediately to simulate network error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client := NewOpenAIClient("test-key", server.URL)
	server.Close() // Close immediately

	_, err := client.ChatCompletion(context.Background(), ChatRequest{Model: "gpt-4"})
	if err == nil {
		t.Error("expected network error, got nil")
	}
}
