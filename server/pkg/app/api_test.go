// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/storage/sqlite"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestAPI_Services(t *testing.T) {
	// Use in-memory SQLite for testing
	db, err := sqlite.NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()
	store := sqlite.NewStore(db)

	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.configPaths = []string{"/config.yaml"}
	// Create dummy config file so ReloadConfig doesn't fail on file not found
	err = afero.WriteFile(app.fs, "/config.yaml", []byte("upstream_services: []"), 0644)
	require.NoError(t, err)

	handler := app.createAPIHandler(store)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := srv.Client()

	t.Run("Create Service", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			Id:   proto.String("test-id"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8080"),
				},
			},
		}
		body, err := protojson.Marshal(svc)
		require.NoError(t, err)

		resp, err := client.Post(srv.URL+"/services", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify it exists in store
		saved, err := store.GetService("test-service")
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, "test-service", saved.GetName())

		httpSvc := saved.GetHttpService()
		require.NotNil(t, httpSvc)
		assert.Equal(t, "http://localhost:8080", httpSvc.GetAddress())
	})

	t.Run("List Services", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/services")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		require.NoError(t, err)
		bodyStr := buf.String()

		assert.Contains(t, bodyStr, `"name":"test-service"`)
		assert.Contains(t, bodyStr, "[")
		assert.Contains(t, bodyStr, "]")
	})

	t.Run("Get Service", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/services/test-service")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		require.NoError(t, err)
		bodyStr := buf.String()
		assert.Contains(t, bodyStr, `"name":"test-service"`)
	})

	t.Run("Get Service Not Found", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/services/non-existent")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Update Service", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:9090"), // Changed port
				},
			},
		}
		body, err := protojson.Marshal(svc)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, srv.URL+"/services/test-service", bytes.NewReader(body))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		saved, err := store.GetService("test-service")
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:9090", saved.GetHttpService().GetAddress())
	})

	t.Run("Delete Service", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/services/test-service", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		saved, err := store.GetService("test-service")
		require.NoError(t, err)
		assert.Nil(t, saved)
	})

	t.Run("Delete Service Missing Name", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/services/", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Method Not Allowed on /services", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, srv.URL+"/services", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Method Not Allowed on /services/{name}", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, srv.URL+"/services/foo", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Create Service Missing Name", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://localhost:8080"),
				},
			},
		}
		body, err := protojson.Marshal(svc)
		require.NoError(t, err)

		resp, err := client.Post(srv.URL+"/services", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create Service Bad JSON", func(t *testing.T) {
		resp, err := client.Post(srv.URL+"/services", "application/json", bytes.NewReader([]byte("{invalid-json")))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Update Service Bad JSON", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/services/foo", bytes.NewReader([]byte("{invalid-json")))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		assert.Equal(t, "OK", buf.String())
	})
}
