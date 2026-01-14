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

func TestContextOptimizerMiddleware_RestoreWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	opt := NewContextOptimizer(10)
	r := gin.New()

	// Upstream middleware
	r.Use(func(c *gin.Context) {
		c.Next()
		// Verify c.Writer is restored
		// Since we cannot easily check type without importing internal/private types or checking for pointer equality with original
		// We rely on the fact that if it was NOT restored, it would be *responseBuffer which is defined in this package.
        // So we can check if it is NOT *responseBuffer.

        // Use reflection or type switch? responseBuffer is exported? No, it's lower case.
        // But we are in the same package (middleware), so we can see responseBuffer.

        _, isResponseBuffer := c.Writer.(*responseBuffer)
        assert.False(t, isResponseBuffer, "Writer should NOT be responseBuffer after middleware returns")
	})

	r.Use(opt.Middleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
