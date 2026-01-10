// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCreateAPIHandler(t *testing.T) {
	// Setup
	store := memory.NewStore()
	app := NewApplication()
	app.fs = afero.NewMemMapFs() // Use in-memory FS to avoid side effects in ReloadConfig
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

		resp, err := http.Post(ts.URL+"/services", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify it was stored
		stored, err := store.GetService(context.Background(), "test-service")
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
		stored, err := store.GetService(context.Background(), "test-service")
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
		stored, err := store.GetService(context.Background(), "test-service")
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
		resp, err = http.Post(ts.URL+"/services", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Memory store creates/overwrites, so it might return 201.
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestHandleServiceStatus(t *testing.T) {
	store := memory.NewStore()
	app := NewApplication()

	// Pre-populate store
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}
	_ = store.SaveService(context.Background(), svc)

	t.Run("get status active", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "test-service", store)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "test-service", resp["name"])
		assert.Equal(t, "Inactive", resp["status"])
	})

	t.Run("service not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/unknown-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "unknown-service", store)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/test-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "test-service", store)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleCollectionApply(t *testing.T) {
	store := memory.NewStore()
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	// No configPaths set to avoid reload failure logs cluttering, but ReloadConfig might fail regardless.
	// We can ignore the reload error as we test the handler logic.

	// Create a collection with one service
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("collection-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}
	collection := &configv1.UpstreamServiceCollectionShare{
		Name: proto.String("test-collection"),
		Services: []*configv1.UpstreamServiceConfig{svc},
	}
	_ = store.SaveServiceCollection(context.Background(), collection)

	t.Run("apply collection success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/test-collection/apply", nil)
		w := httptest.NewRecorder()

		app.handleCollectionApply(w, req, "test-collection", store)

		assert.Equal(t, http.StatusOK, w.Code)

		savedSvc, err := store.GetService(context.Background(), "collection-service")
		require.NoError(t, err)
		require.NotNil(t, savedSvc)
		assert.Equal(t, "collection-service", savedSvc.GetName())
	})

	t.Run("apply collection not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/unknown-collection/apply", nil)
		w := httptest.NewRecorder()

		app.handleCollectionApply(w, req, "unknown-collection", store)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/collections/test-collection/apply", nil)
		w := httptest.NewRecorder()

		app.handleCollectionApply(w, req, "test-collection", store)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleSettings(t *testing.T) {
	store := memory.NewStore()
	app := NewApplication()
	app.fs = afero.NewMemMapFs()

	t.Run("get settings default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		w := httptest.NewRecorder()

		handler := app.handleSettings(store)
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// It returns default values, not empty JSON.
		assert.Contains(t, w.Body.String(), `"allowed_ips":[]`)
	})

	t.Run("save settings", func(t *testing.T) {
		settingsJSON := `{"allowed_ips": ["127.0.0.1"]}`
		req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewBufferString(settingsJSON))
		w := httptest.NewRecorder()

		handler := app.handleSettings(store)
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify saved
		settings, err := store.GetGlobalSettings()
		require.NoError(t, err)
		assert.Contains(t, settings.AllowedIps, "127.0.0.1")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/settings", nil)
		w := httptest.NewRecorder()

		handler := app.handleSettings(store)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
