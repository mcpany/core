// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func setupTemplateTestApp(t *testing.T) *Application {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	// Clear seeded templates for testing
	tm.templates = []*configv1.UpstreamServiceConfig{}
	_ = tm.save() // Ensure empty state is saved
	return &Application{
		TemplateManager: tm,
	}
}

func TestHandleTemplates_List(t *testing.T) {
	app := setupTemplateTestApp(t)
	handler := app.handleTemplates()

	t.Run("Empty List", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, "[]", w.Body.String())
	})

	t.Run("Populated List", func(t *testing.T) {
		// Add a template directly
		tmpl := &configv1.UpstreamServiceConfig{}
		tmpl.SetName("Test Template")
		tmpl.SetId("test-id")
		require.NoError(t, app.TemplateManager.SaveTemplate(tmpl))

		req := httptest.NewRequest(http.MethodGet, "/templates", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var list []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &list)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Test Template", list[0]["name"])
	})
}

func TestHandleTemplates_Create(t *testing.T) {
	app := setupTemplateTestApp(t)
	handler := app.handleTemplates()

	t.Run("Success", func(t *testing.T) {
		tmpl := &configv1.UpstreamServiceConfig{}
		tmpl.SetName("New Template")

		opts := protojson.MarshalOptions{UseProtoNames: true}
		body, err := opts.Marshal(tmpl)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp["id"])

		// Verify it was saved
		saved := app.TemplateManager.ListTemplates()
		assert.Len(t, saved, 1)
		assert.Equal(t, "New Template", saved[0].GetName())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/templates", strings.NewReader("{invalid-json"))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Body Too Large", func(t *testing.T) {
		// Create a large body > 1MB
		largeBody := make([]byte, 2*1024*1024)
		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(largeBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// MaxBytesReader typically returns an error on Read, which protojson.Unmarshal catches
		// The error message from protojson might vary, but status code should be 400
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleTemplates_MethodNotAllowed(t *testing.T) {
	app := setupTemplateTestApp(t)
	handler := app.handleTemplates()

	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/templates", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

func TestHandleTemplateDetail_Delete(t *testing.T) {
	app := setupTemplateTestApp(t)
	handler := app.handleTemplateDetail()

	t.Run("Success", func(t *testing.T) {
		// Setup existing template
		id := uuid.New().String()
		tmpl := &configv1.UpstreamServiceConfig{}
		tmpl.SetName("To Delete")
		tmpl.SetId(id)

		require.NoError(t, app.TemplateManager.SaveTemplate(tmpl))

		req := httptest.NewRequest(http.MethodDelete, "/templates/"+id, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, app.TemplateManager.ListTemplates())
	})

	t.Run("Not Found (Idempotent)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/templates/non-existent", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		// The routing usually handles this, but the handler logic has a check
		req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleTemplateDetail_MethodNotAllowed(t *testing.T) {
	app := setupTemplateTestApp(t)
	handler := app.handleTemplateDetail()

	// GET and PUT are not implemented for detail in the current code
	methods := []string{http.MethodGet, http.MethodPut, http.MethodPost}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/templates/some-id", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}
