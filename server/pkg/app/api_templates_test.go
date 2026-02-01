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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleTemplates_Get(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}
	handler := app.handleTemplates()

	// Seed a template
	tmpl := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-template"),
		Id:   proto.String("test-id"),
	}.Build()
	require.NoError(t, tm.SaveTemplate(tmpl))

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var templates []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &templates)
	require.NoError(t, err)

	// Check if test-template is in the list
	found := false
	for _, t := range templates {
		if t["name"] == "test-template" && t["id"] == "test-id" {
			found = true
			break
		}
	}
	assert.True(t, found, "test-template not found in response")
}

func TestHandleTemplates_Post(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}
	handler := app.handleTemplates()

	t.Run("Valid Template", func(t *testing.T) {
		tmpl := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("new-template"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("http://localhost"),
			}.Build(),
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		body, err := opts.Marshal(tmpl)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify saved
		saved := tm.ListTemplates()
		found := false
		for _, s := range saved {
			if s.GetName() == "new-template" {
				found = true
				assert.NotEmpty(t, s.GetId())
				break
			}
		}
		assert.True(t, found, "new-template not found in manager")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("{invalid-json")))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleTemplateDetail_Delete(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}
	handler := app.handleTemplateDetail()

	// Seed a template
	tmpl := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("to-delete"),
		Id:   proto.String("delete-id"),
	}.Build()
	require.NoError(t, tm.SaveTemplate(tmpl))

	t.Run("Delete Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/templates/delete-id", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deleted
		saved := tm.ListTemplates()
		found := false
		for _, s := range saved {
			if s.GetId() == "delete-id" {
				found = true
				break
			}
		}
		assert.False(t, found, "template delete-id should have been deleted")
	})

	t.Run("Delete Non-Existent", func(t *testing.T) {
		// Deleting non-existent is idempotent in manager usually, verifying response code
		// The manager implementation:
		// func (tm *TemplateManager) DeleteTemplate(idOrName string) error { ... }
		// It iterates and filters. It returns nil even if not found (idempotent).

		req := httptest.NewRequest(http.MethodDelete, "/templates/non-existent", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestHandleTemplateDetail_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	app := &Application{TemplateManager: tm}
	handler := app.handleTemplateDetail()

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/templates/some-id", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		// The handler trims prefix "/templates/". If we send just "/templates/", id is empty.
		req := httptest.NewRequest(http.MethodDelete, "/templates/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "id required")
	})
}
