// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleServiceDiagnose(t *testing.T) {
	// Allow loopback for tests
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

	// Setup mock upstream
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Setup App and Store
	app, store := setupApiTestApp() // reuse helper from api_test.go
	// Note: setupApiTestApp is in api_test.go, which is in same package 'app'.
	// However, if api_test.go is not compiled in the same test run (it is if I run go test ./pkg/app), it works.

	// Save a service pointing to mock upstream
	svcName := "test-doctor-service"
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(svcName),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(ts.URL),
		}.Build(),
	}.Build()

	// We need to ensure store is populated.
	// Since setupApiTestApp returns a memory store which implements Storage, we can use it.
	err := store.SaveService(context.Background(), svc)
	require.NoError(t, err)

	handler := app.handleServiceDiagnose(store)

	t.Run("Diagnose Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/"+svcName+"/diagnose", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, svcName, resp["service_name"])
		assert.Equal(t, "OK", resp["status"])
		assert.Contains(t, resp["message"], "Service reachable")
	})

	t.Run("Diagnose Service Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/unknown-service/diagnose", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/"+svcName+"/diagnose", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
