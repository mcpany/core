// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextOptimizerMiddleware(t *testing.T) {
	opt := NewContextOptimizer(10)

	mux := http.NewServeMux()
	mux.HandleFunc("/long", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "This text is definitely longer than 10 characters",
					},
				},
			},
		})
	})

	mux.HandleFunc("/short", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Short",
					},
				},
			},
		})
	})

	handler := opt.Handler(mux)

	// Test Long Response
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/long", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	result := resp["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)

	assert.True(t, strings.Contains(text, "TRUNCATED"), "Should contain truncated message")
	assert.True(t, len(text) < 50, "Should be truncated")

	// Test Short Response
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/short", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	result = resp["result"].(map[string]interface{})
	content = result["content"].([]interface{})
	text = content[0].(map[string]interface{})["text"].(string)

	assert.Equal(t, "Short", text)
}

func TestContextOptimizerMiddleware_WriterType(t *testing.T) {
	opt := NewContextOptimizer(10)

	var isResponseBuffer bool
	checkHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, isResponseBuffer = w.(*responseBuffer)
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "pong"}`))
	})

	handler := opt.Handler(checkHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	handler.ServeHTTP(w, req)

	assert.True(t, isResponseBuffer, "Inner handler should receive responseBuffer")
	assert.Equal(t, 200, w.Code)
}

func TestContextOptimizerMiddleware_MultiByteTruncation(t *testing.T) {
	// Set MaxChars to 1.
	// We will send a string starting with an Emoji (4 bytes).
	// If the implementation uses byte slicing, str[:1] will be invalid UTF-8.
	opt := NewContextOptimizer(1)

	mux := http.NewServeMux()
	mux.HandleFunc("/emoji", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 🚀 is 4 bytes: F0 9F 9A 80
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "🚀 Rocket",
					},
				},
			},
		})
	})

	handler := opt.Handler(mux)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/emoji", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Response body should be valid JSON even after truncation")

	result := resp["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)

	// Ensure the text starts with valid UTF-8 and contains the truncation message.
	assert.True(t, strings.Contains(text, "TRUNCATED"), "Should contain truncated message")
	// If it was corrupted, Unmarshal might have failed or replaced with replacement char.
	// But Go json decoder might replace invalid bytes with \ufffd.
	// We want to ensure it is NOT \ufffd.
	assert.False(t, strings.Contains(text, "\ufffd"), "Should not contain replacement character for invalid UTF-8")
}
