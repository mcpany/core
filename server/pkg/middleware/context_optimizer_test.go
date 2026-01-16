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
)

func TestContextOptimizerMiddleware(t *testing.T) {
	opt := NewContextOptimizer(10)

	// Handler that returns a long text
	longHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"result": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "This text is definitely longer than 10 characters",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Handler that returns short text
	shortHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"result": map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Short",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Wrap handlers
	wrappedLong := opt.Handler(longHandler)
	wrappedShort := opt.Handler(shortHandler)

	// Test Long Response
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/long", nil)
	wrappedLong.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	result := resp["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)

	assert.True(t, strings.Contains(text, "TRUNCATED"), "Should contain truncated message")
	assert.True(t, len(text) < 50, "Should be truncated")

	// Test Short Response
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/short", nil)
	wrappedShort.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	result = resp["result"].(map[string]interface{})
	content = result["content"].([]interface{})
	text = content[0].(map[string]interface{})["text"].(string)

	assert.Equal(t, "Short", text)
}
