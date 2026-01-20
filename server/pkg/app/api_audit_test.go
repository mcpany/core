// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleAuditExport(t *testing.T) {
	app := setupTestApp()
	app.standardMiddlewares = &middleware.StandardMiddlewares{}

	// 1. Setup real SQLite audit store in CWD (allowed path)
	dbPath := "./audit_test_export.db"
	defer os.Remove(dbPath)

	// Pre-create some data
	sqliteStore, err := middleware.NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)

	entry1 := middleware.AuditEntry{
		Timestamp:  time.Now().Add(-1 * time.Hour),
		ToolName:   "tool-1",
		UserID:     "user-1",
		DurationMs: 123,
		Arguments:  []byte(`{"key":"val"}`),
	}
	require.NoError(t, sqliteStore.Write(context.Background(), entry1))
	sqliteStore.Close()

	// Initialize AuditMiddleware with the SQLite store
	storageType := configv1.AuditConfig_STORAGE_TYPE_SQLITE
	audit, err := middleware.NewAuditMiddleware(&configv1.AuditConfig{
		Enabled:     proto.Bool(true),
		StorageType: &storageType,
		OutputPath:  proto.String(dbPath),
	})
	require.NoError(t, err)
	app.standardMiddlewares.Audit = audit
	defer audit.Close()

	// 2. Test Export
	req, _ := http.NewRequest("GET", "/api/v1/audit/export?tool_name=tool-1", nil)
	rr := httptest.NewRecorder()

	mux := app.createAPIHandler(app.Storage)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))

	csvReader := csv.NewReader(rr.Body)
	records, err := csvReader.ReadAll()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(records), 2) // Header + 1 row
	assert.Equal(t, "tool-1", records[1][1])
	assert.Equal(t, "user-1", records[1][2])
	assert.Equal(t, `{"key":"val"}`, records[1][4])
}
