// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDebuggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	debugger := NewDebugger(10)
	r := gin.New()
	r.Use(debugger.Middleware())

	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check entries
	entries := debugger.Entries()
	assert.Len(t, entries, 1)
	assert.Equal(t, "/test", entries[0].Path)
	assert.Equal(t, http.StatusOK, entries[0].Status)

	// Test Ring Buffer Overflow
	for i := 0; i < 15; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
	}

	entries = debugger.Entries()
	assert.Len(t, entries, 10) // Should only keep last 10
}
