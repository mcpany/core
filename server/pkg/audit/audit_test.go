// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresAuditStore_Close(t *testing.T) {
	// Create a mock DB
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() // In real usage, store.Close() will close it.

	mock.ExpectClose()

	store := &PostgresAuditStore{
		db: db,
	}

	err = store.Close()
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWebhookAuditStore_Close(t *testing.T) {
	store := &WebhookAuditStore{}
	err := store.Close()
	require.NoError(t, err)
}

func TestWebhookAuditStore_Write(t *testing.T) {
	// Setup a mock server
	done := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webhook", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "bar", r.Header.Get("X-Foo"))

		var batch []Entry
		_ = json.NewDecoder(r.Body).Decode(&batch)
		if len(batch) > 0 && batch[0].ToolName == "test" {
			close(done)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	store := NewWebhookAuditStore(ts.URL+"/webhook", map[string]string{"X-Foo": "bar"})
	err := store.Write(context.Background(), Entry{ToolName: "test"})
	require.NoError(t, err)

	_ = store.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for webhook")
	}
}

func TestWebhookAuditStore_Write_Error(t *testing.T) {
	// Webhook Write is now async, so it won't return errors from the server.
	// We just verify it doesn't panic and we can close it.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	store := NewWebhookAuditStore(ts.URL, nil)
	err := store.Write(context.Background(), Entry{ToolName: "test"})
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)
}
