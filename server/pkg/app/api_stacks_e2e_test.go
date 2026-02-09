// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TestStackLifecycle tests the full lifecycle of a stack (collection): Create -> Apply -> Verify Service.
func TestStackLifecycle(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.createAPIHandler(store)

	stackName := "e2e-stack"
	serviceName := "e2e-service"

	// 1. Create Stack (Collection)
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceName),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:9090"),
		}.Build(),
	}.Build()

	collection := configv1.Collection_builder{
		Name:        proto.String(stackName),
		Description: proto.String("End-to-end test stack"),
		Services:    []*configv1.UpstreamServiceConfig{svc},
	}.Build()

	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(collection)
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// 2. Verify Stack Exists
	req = httptest.NewRequest(http.MethodGet, "/collections/"+stackName, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 3. Apply Stack
	req = httptest.NewRequest(http.MethodPost, "/collections/"+stackName+"/apply", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 4. Verify Service Created (Side Effect)
	// We check the store directly or via API
	req = httptest.NewRequest(http.MethodGet, "/services/"+serviceName, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var createdSvc configv1.UpstreamServiceConfig
	err := protojson.Unmarshal(w.Body.Bytes(), &createdSvc)
	require.NoError(t, err)
	assert.Equal(t, serviceName, createdSvc.GetName())
	assert.Equal(t, "http://127.0.0.1:9090", createdSvc.GetHttpService().GetAddress())
}
