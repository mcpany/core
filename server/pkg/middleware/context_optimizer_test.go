// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestContextOptimizerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	opt := NewContextOptimizer(10)
	r := gin.New()
	r.Use(opt.Middleware())

	r.POST("/long", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
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

	r.POST("/short", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
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

	// Test Long Response
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/long", nil)
	r.ServeHTTP(w, req)

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
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	result = resp["result"].(map[string]interface{})
	content = result["content"].([]interface{})
	text = content[0].(map[string]interface{})["text"].(string)

	assert.Equal(t, "Short", text)
}
