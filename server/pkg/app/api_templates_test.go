// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string {
	return &s
}

func TestHandleTemplates(t *testing.T) {
	// Setup Application with a real TemplateManager using a temp dir
	tempDir := t.TempDir()
	app := &Application{
		TemplateManager: NewTemplateManager(tempDir),
	}

	handler := app.handleTemplates()

	t.Run("ListTemplates_Empty", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/templates", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var templates []map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &templates)
		require.NoError(t, err)
		assert.Empty(t, templates)
	})

	t.Run("CreateTemplate", func(t *testing.T) {
		template := map[string]interface{}{
			"name": "test-template",
			"id":   "test-id",
			"mcp_service": map[string]interface{}{
				"http_connection": map[string]interface{}{
					"http_address": "http://localhost:8080",
				},
			},
		}
		bodyBytes, _ := json.Marshal(template)

		req, err := http.NewRequest(http.MethodPost, "/templates", bytes.NewReader(bodyBytes))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]string
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "test-id", resp["id"])

		// Verify it was stored
		stored := app.TemplateManager.ListTemplates()
		assert.Len(t, stored, 1)
		assert.Equal(t, "test-template", stored[0].GetName())
	})

	t.Run("CreateTemplate_InvalidJSON", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("{invalid-json")))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "/templates", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestHandleTemplateDetail(t *testing.T) {
	// Setup Application with a real TemplateManager using a temp dir
	tempDir := t.TempDir()
	app := &Application{
		TemplateManager: NewTemplateManager(tempDir),
	}

	// Add a template to delete
	tmpl := &configv1.UpstreamServiceConfig{
		Name: stringPtr("test-template"),
		Id:   stringPtr("test-id"),
	}
	err := app.TemplateManager.SaveTemplate(tmpl)
	require.NoError(t, err)

	handler := app.handleTemplateDetail()

	t.Run("DeleteTemplate", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/templates/test-id", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify it was deleted
		stored := app.TemplateManager.ListTemplates()
		assert.Empty(t, stored)
	})

	t.Run("DeleteTemplate_MissingID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/templates/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/templates/test-id", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
