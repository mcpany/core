// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockStore struct {
	entries    []audit.Entry
	lastFilter audit.Filter
}

func (m *mockStore) Write(ctx context.Context, entry audit.Entry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	m.lastFilter = filter
	return m.entries, nil
}

func (m *mockStore) Close() error {
	return nil
}

func TestExportAuditLogsHandler(t *testing.T) {
	// Setup
	config := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()

	mw, _ := middleware.NewAuditMiddleware(config)
	store := &mockStore{
		entries: []audit.Entry{
			{
				Timestamp: time.Now(),
				ToolName:  "test-tool",
				UserID:    "test-user",
				Duration:  "100ms",
			},
		},
	}
	mw.SetStore(store)

	handler := ExportAuditLogsHandler(mw)

	t.Run("Default Parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/audit/export", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
		assert.Equal(t, 10000, store.lastFilter.Limit, "Default limit should be 10000")
	})

	t.Run("Custom Parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/audit/export?tool_name=mytool&limit=5&user_id=alice", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "mytool", store.lastFilter.ToolName)
		assert.Equal(t, "alice", store.lastFilter.UserID)
		assert.Equal(t, 5, store.lastFilter.Limit)
	})

	t.Run("Limit Cap", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/audit/export?limit=999999", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, 10000, store.lastFilter.Limit, "Limit should be capped at 10000")
	})
}
