// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuditStore struct {
	mock.Mock
}

func (m *MockAuditStore) Write(ctx context.Context, entry audit.Entry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockAuditStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]audit.Entry), args.Error(1)
}

func (m *MockAuditStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestHandleAuditExport_Mock(t *testing.T) {
	app := NewApplication()
	mockStore := new(MockAuditStore)

	// Initialize middleware
	auditConfig := &configv1.AuditConfig{}
	auditConfig.SetEnabled(true)
	am, err := middleware.NewAuditMiddleware(auditConfig)
	require.NoError(t, err)
	am.SetStore(mockStore)

	app.standardMiddlewares = &middleware.StandardMiddlewares{
		Audit: am,
	}

	t.Run("Success", func(t *testing.T) {
		entries := []audit.Entry{
			{
				Timestamp:  time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				ToolName:   "test-tool",
				UserID:     "user-1",
				ProfileID:  "profile-1",
				DurationMs: 100,
				Arguments:  []byte(`{"arg":"val"}`),
				Result:     "success",
			},
		}
		mockStore.On("Read", mock.Anything, mock.Anything).Return(entries, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/audit/export", nil)
		w := httptest.NewRecorder()

		app.handleAuditExport(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))

		csvReader := csv.NewReader(w.Body)
		records, err := csvReader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, 2, len(records)) // Header + 1 row
		assert.Equal(t, "test-tool", records[1][1])
	})

	t.Run("StoreError", func(t *testing.T) {
		mockStore.On("Read", mock.Anything, mock.Anything).Return([]audit.Entry{}, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/audit/export", nil)
		w := httptest.NewRecorder()

		app.handleAuditExport(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("NotConfigured", func(t *testing.T) {
		app.standardMiddlewares.Audit = nil
		req := httptest.NewRequest(http.MethodGet, "/audit/export", nil)
		w := httptest.NewRecorder()

		app.handleAuditExport(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		// Restore middleware
		app.standardMiddlewares.Audit = am
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/audit/export", nil)
		w := httptest.NewRecorder()
		app.handleAuditExport(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleAuditLogs(t *testing.T) {
	app, _ := setupApiTestApp()
	app.standardMiddlewares = &middleware.StandardMiddlewares{}

	// Use CWD to ensure path is allowed by validation
	dbPath := "./audit_test_logs.db"
	defer os.Remove(dbPath)

	// Allow .db extension for this test as IsSensitivePath blocks it by default
	originalIsSensitive := validation.IsSensitivePath
	validation.IsSensitivePath = func(path string) error {
		if strings.HasSuffix(path, "audit_test_logs.db") {
			return nil
		}
		return originalIsSensitive(path)
	}
	defer func() { validation.IsSensitivePath = originalIsSensitive }()

	sqliteStore, err := audit.NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	entry1 := audit.Entry{
		Timestamp:  time.Now().Add(-1 * time.Hour),
		ToolName:   "tool-1",
		UserID:     "user-1",
		DurationMs: 123,
		Arguments:  []byte(`{"key":"val"}`),
	}
	require.NoError(t, sqliteStore.Write(context.Background(), entry1))
	sqliteStore.Close()

	storageType := configv1.AuditConfig_STORAGE_TYPE_SQLITE
	auditCfg := &configv1.AuditConfig{}
	auditCfg.SetEnabled(true)
	auditCfg.SetStorageType(storageType)
	auditCfg.SetOutputPath(dbPath)
	auditMiddleware, err := middleware.NewAuditMiddleware(auditCfg)
	require.NoError(t, err)
	app.standardMiddlewares.Audit = auditMiddleware
	defer auditMiddleware.Close()

	req, _ := http.NewRequest("GET", "/audit/logs?tool_name=tool-1", nil)
	rr := httptest.NewRecorder()
	// Use createAPIHandler to verify routing
	mux := app.createAPIHandler(app.Storage)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string][]audit.Entry
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	entries := resp["entries"]
	assert.Len(t, entries, 1)
	assert.Equal(t, "tool-1", entries[0].ToolName)
	assert.Equal(t, "user-1", entries[0].UserID)
}
