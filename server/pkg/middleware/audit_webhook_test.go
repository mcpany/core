// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookAuditStore(t *testing.T) {
	// Setup test server
	var receivedEntry AuditEntry
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		err := json.NewDecoder(r.Body).Decode(&receivedEntry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup store
	headers := map[string]string{
		"X-Custom-Header": "test-value",
	}
	store := NewWebhookAuditStore(server.URL, headers)

	// Create test entry
	entry := AuditEntry{
		Timestamp:   time.Now(),
		ToolName:    "test-tool",
		UserID:      "test-user",
		DurationMs:  100,
		Arguments:   json.RawMessage(`{"arg": "value"}`),
		Result:      "success",
		Error:       "",
	}

	// Test Write
	err := store.Write(context.Background(), entry)
	require.NoError(t, err)

	// Verify received data
	assert.Equal(t, entry.ToolName, receivedEntry.ToolName)
	assert.Equal(t, entry.UserID, receivedEntry.UserID)
	assert.Equal(t, entry.DurationMs, receivedEntry.DurationMs)
	assert.JSONEq(t, string(entry.Arguments), string(receivedEntry.Arguments))
	assert.Equal(t, entry.Result, receivedEntry.Result)

	// Verify headers
	assert.Equal(t, "test-value", receivedHeaders.Get("X-Custom-Header"))
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
}

func TestWebhookAuditStore_Failure(t *testing.T) {
	// Setup test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	store := NewWebhookAuditStore(server.URL, nil)
	entry := AuditEntry{ToolName: "test-tool"}

	err := store.Write(context.Background(), entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook returned status: 500")
}
