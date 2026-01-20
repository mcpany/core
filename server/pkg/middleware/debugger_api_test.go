// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerAPIHandler(t *testing.T) {
	debugger := NewDebugger(10)

	// Populate some entries
	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "value"}`))
	}))

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"input": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Get the ID of the entry
	entries := debugger.Entries()
	assert.Len(t, entries, 1)
	entryID := entries[0].ID

	apiHandler := debugger.APIHandler()

	t.Run("ListAll", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/debug/entries", nil)
		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []DebugEntry
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.NotEmpty(t, resp[0].RequestBody)
		assert.NotEmpty(t, resp[0].ResponseBody)
	})

	t.Run("ListSummary", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/debug/entries?summary=true", nil)
		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []DebugEntry
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Empty(t, resp[0].RequestBody)
		assert.Empty(t, resp[0].ResponseBody)
		assert.NotEmpty(t, resp[0].ID)
	})

	t.Run("GetByID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/debug/entries?id="+entryID, nil)
		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp DebugEntry
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, entryID, resp.ID)
		assert.NotEmpty(t, resp.RequestBody)
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/debug/entries?id=invalid", nil)
		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
