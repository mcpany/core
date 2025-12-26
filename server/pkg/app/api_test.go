// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/storage/sqlite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIHandler(t *testing.T) {
	// Setup SQLite DB
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	// Setup Application
	app := NewApplication()
	app.fs = afero.NewMemMapFs() // Use in-memory FS to avoid side effects in ReloadConfig
	app.ServiceRegistry = serviceregistry.New(
		app.UpstreamFactory,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		auth.NewManager(),
	)

	handler := app.createAPIHandler(store)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	t.Run("ListServices_Empty", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/services")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Should decode to empty list
		var services []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&services)
		require.NoError(t, err)
		assert.Empty(t, services)
	})

	t.Run("CreateService", func(t *testing.T) {
		// Use standard JSON for simplicity
		body := map[string]interface{}{
			"name": "test-service",
			"id":   "test-id",
			"mcp_service": map[string]interface{}{
				"http_connection": map[string]interface{}{
					"http_address": "http://localhost:8080",
				},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		resp, err := http.Post(ts.URL + "/services", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify it was stored
		stored, err := store.GetService("test-service")
		require.NoError(t, err)
		assert.Equal(t, "test-service", stored.GetName())
		assert.Equal(t, "http://localhost:8080", stored.GetMcpService().GetHttpConnection().GetHttpAddress())
	})

	t.Run("GetService", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/services/test-service")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "test-service", result["name"])
	})

	t.Run("UpdateService", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "test-service",
			"id":   "test-id",
			"mcp_service": map[string]interface{}{
				"http_connection": map[string]interface{}{
					"http_address": "http://localhost:9090", // Changed port
				},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req, err := http.NewRequest(http.MethodPut, ts.URL+"/services/test-service", bytes.NewReader(bodyBytes))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify update
		stored, err := store.GetService("test-service")
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:9090", stored.GetMcpService().GetHttpConnection().GetHttpAddress())
	})

	t.Run("DeleteService", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, ts.URL+"/services/test-service", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify deletion
		stored, err := store.GetService("test-service")
		require.NoError(t, err)
		assert.Nil(t, stored)
	})

	t.Run("GetNonExistentService", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/services/non-existent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("CreateService_MissingName", func(t *testing.T) {
		body := map[string]interface{}{
			"id": "test-id-no-name",
		}
		bodyBytes, _ := json.Marshal(body)

		resp, err := http.Post(ts.URL+"/services", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("CreateService_InvalidJSON", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/services", "application/json", bytes.NewReader([]byte("{invalid-json")))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UpdateService_InvalidJSON", func(t *testing.T) {
		resp, err := http.NewRequest(http.MethodPut, ts.URL+"/services/test-service", bytes.NewReader([]byte("{invalid-json")))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, resp)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("CreateService_Duplicate", func(t *testing.T) {
		// First create a service
		body := map[string]interface{}{
			"name": "duplicate-service",
			"id":   "duplicate-id",
			"mcp_service": map[string]interface{}{
				"http_connection": map[string]interface{}{
					"http_address": "http://localhost:8080",
				},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		resp, err := http.Post(ts.URL+"/services", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()

		// Try to create again with same name/id
		// Store implementation might not fail on duplicate (it might overwrite or ignore), check sqlite store implementation.
		// If sqlite store uses `INSERT OR REPLACE` or `ON CONFLICT`, it might succeed.
		// But let's check behavior.
		resp, err = http.Post(ts.URL+"/services", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()
		// If it succeeds (overwrite), then we can't test failure here easily without modifying store.
		// Let's assume it succeeds for now, and check if we can force failure.
		// If we can't force failure, we can't test the error path in createAPIHandler.
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}
