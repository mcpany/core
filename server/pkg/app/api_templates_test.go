// Copyright 2026 Author(s) of MCP Any
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleTemplates(t *testing.T) {
	app, _ := setupApiTestApp()

	// 1. Test GET (Empty)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/service-templates", nil)
	w := httptest.NewRecorder()
	app.handleTemplates().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// We need to decode into a struct that matches the JSON response
	// Since the response is a list of ServiceTemplate messages
	var list []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.Empty(t, list)

	// 2. Test POST (Create)
	tmpl := configv1.ServiceTemplate_builder{
		Id:          proto.String("test-tmpl"),
		Name:        proto.String("Test Template"),
		Description: proto.String("A test template"),
		ServiceConfig: configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("http://example.com"),
			}.Build(),
		}.Build(),
	}.Build()

	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(tmpl)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/service-templates", bytes.NewReader(body))
	w = httptest.NewRecorder()
	app.handleTemplates().ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 3. Test GET (Verify Created)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/service-templates", nil)
	w = httptest.NewRecorder()
	app.handleTemplates().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "test-tmpl", list[0]["id"])
	assert.Equal(t, "Test Template", list[0]["name"])
}

func TestHandleTemplateDetail(t *testing.T) {
	app, store := setupApiTestApp()

	// Seed a template
	tmpl := configv1.ServiceTemplate_builder{
		Id:   proto.String("test-tmpl"),
		Name: proto.String("Test Template"),
	}.Build()
	require.NoError(t, store.SaveServiceTemplate(context.Background(), tmpl))

	// 1. Test DELETE
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/service-templates/test-tmpl", nil)
	w := httptest.NewRecorder()
	app.handleTemplateDetail().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// 2. Verify Deleted
	list, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.Empty(t, list)

	// 3. Test DELETE Non-Existent (Idempotent, so 204 is acceptable/expected)
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/service-templates/non-existent", nil)
	w = httptest.NewRecorder()
	app.handleTemplateDetail().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
