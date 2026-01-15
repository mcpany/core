// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleDiagnose(t *testing.T) {
	// Setup SQLite DB
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	// Setup Application
	app := NewApplication()
	app.fs = afero.NewMemMapFs()

	// Seed a service in the store
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-diag-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://example.com"),
			},
		},
	}
	err = store.SaveService(context.Background(), svc)
	require.NoError(t, err)

	handler := app.createAPIHandler(store)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	t.Run("Diagnose_Success", func(t *testing.T) {
		reqBody := map[string]string{
			"service_name": "test-diag-service",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := http.Post(ts.URL+"/diagnose", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var report map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&report)
		require.NoError(t, err)

		assert.Equal(t, "test-diag-service", report["service_name"])
		steps, ok := report["steps"].([]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, steps)
	})

	t.Run("Diagnose_NotFound", func(t *testing.T) {
		reqBody := map[string]string{
			"service_name": "non-existent-service",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := http.Post(ts.URL+"/diagnose", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Diagnose_BadRequest_MissingName", func(t *testing.T) {
		reqBody := map[string]string{
			"service_name": "",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := http.Post(ts.URL+"/diagnose", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Diagnose_InvalidMethod", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/diagnose")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
