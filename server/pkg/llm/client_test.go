// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	// Mock Server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		// Verify Headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Verify Body
		var req openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.Model != "gpt-4" {
			t.Errorf("Expected model gpt-4, got %s", req.Model)
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "Hello" {
			t.Errorf("Unexpected messages: %v", req.Messages)
		}

		// Response
		resp := openAIChatResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			Message: struct {
				Content string `json:"content"`
			}{
				Content: "Hi there!",
			},
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewOpenAIClient("test-key", mockServer.URL)

	resp, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Content != "Hi there!" {
		t.Errorf("Expected content 'Hi there!', got '%s'", resp.Content)
	}
}
