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

func setupTestAppWithTemplates(t *testing.T) *Application {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	// Clear seeded templates for testing isolation
	tm.templates = nil
	return &Application{
		TemplateManager: tm,
	}
}

func TestHandleTemplates_List(t *testing.T) {
	app := setupTestAppWithTemplates(t)
	handler := app.handleTemplates()

	// 1. Empty List
	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, "[]", w.Body.String())

	// 2. Populated List
	tmpl := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-template"),
		Id:   proto.String("test-id"),
	}.Build()
	require.NoError(t, app.TemplateManager.SaveTemplate(tmpl))

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var list []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "test-template", list[0]["name"])
}

func TestHandleTemplates_Create(t *testing.T) {
	app := setupTestAppWithTemplates(t)
	handler := app.handleTemplates()

	// 1. Success
	tmpl := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("new-template"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://example.com"),
		}.Build(),
	}.Build()

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

	// Verify persistence
	saved := app.TemplateManager.ListTemplates()[0]
	assert.Equal(t, "new-template", saved.GetName())

	// 2. Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte("{invalid-json")))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 3. Valid JSON but unrelated content
	req = httptest.NewRequest(http.MethodPost, "/templates", bytes.NewReader([]byte(`{"unknown_field": "value"}`)))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// protojson unmarshal errors on unknown fields by default
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTemplateDetail_Delete(t *testing.T) {
	app := setupTestAppWithTemplates(t)
	handler := app.handleTemplateDetail()

	// Seed
	tmpl := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("to-delete"),
		Id:   proto.String("del-id"),
	}.Build()
	require.NoError(t, app.TemplateManager.SaveTemplate(tmpl))

	// 1. Success Delete
	req := httptest.NewRequest(http.MethodDelete, "/templates/del-id", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, app.TemplateManager.ListTemplates())

	// 2. Delete Non-Existent (idempotent operation)
	req = httptest.NewRequest(http.MethodDelete, "/templates/non-existent", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandleTemplateDetail_Methods(t *testing.T) {
	app := setupTestAppWithTemplates(t)
	handler := app.handleTemplateDetail()

	// GET Not Allowed
	req := httptest.NewRequest(http.MethodGet, "/templates/some-id", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// POST Not Allowed
	req = httptest.NewRequest(http.MethodPost, "/templates/some-id", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleTemplates_Methods(t *testing.T) {
	app := setupTestAppWithTemplates(t)
	handler := app.handleTemplates()

	// DELETE Not Allowed on collection
	req := httptest.NewRequest(http.MethodDelete, "/templates", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
