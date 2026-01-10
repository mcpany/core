// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
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

	r.POST("/echo", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.BindJSON(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// 1. Test Basic GET
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	entries := debugger.Entries()
	assert.NotEmpty(t, entries)
	lastEntry := entries[len(entries)-1]
	assert.Equal(t, "/test", lastEntry.Path)
	assert.NotEmpty(t, lastEntry.ID)

	// 2. Test POST Body Capture
	w = httptest.NewRecorder()
	reqBody := `{"message": "hello"}`
	req, _ = http.NewRequest("POST", "/echo", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, reqBody, w.Body.String())

	entries = debugger.Entries()
	lastEntry = entries[len(entries)-1]
	assert.Equal(t, "/echo", lastEntry.Path)
	assert.Equal(t, reqBody, lastEntry.RequestBody)
	assert.JSONEq(t, reqBody, lastEntry.ResponseBody)

	// 3. Test Replay Handler
	r.POST("/debug/replay", debugger.ReplayHandler())

	// Start a mock server to replay against
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Replay", "true")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("replayed"))
	}))
	defer ts.Close()

	replayReq := ReplayRequest{
		Method: "POST",
		URL:    ts.URL,
		Body:   "test payload",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}
	replayBody, _ := json.Marshal(replayReq)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/debug/replay", bytes.NewBuffer(replayBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var replayResp ReplayResponse
	err := json.Unmarshal(w.Body.Bytes(), &replayResp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, replayResp.Status)
	assert.Equal(t, "replayed", replayResp.Body)
}
