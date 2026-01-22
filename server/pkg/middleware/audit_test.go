// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webhook", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "bar", r.Header.Get("X-Foo"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	store := NewWebhookAuditStore(ts.URL+"/webhook", map[string]string{"X-Foo": "bar"})
	err := store.Write(context.Background(), AuditEntry{ToolName: "test"})
	require.NoError(t, err)
}

func TestWebhookAuditStore_Write_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	store := NewWebhookAuditStore(ts.URL, nil)
	err := store.Write(context.Background(), AuditEntry{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook returned status")
}
