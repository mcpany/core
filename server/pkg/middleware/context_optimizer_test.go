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
