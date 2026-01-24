// Copyright 2025 Author(s) of MCP Any
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

func TestHandleServiceValidate_MethodNotAllowed(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodGet, "/services/validate", nil)
	w := httptest.NewRecorder()
	app.handleServiceValidate().ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleServiceValidate_MalformedJSON(t *testing.T) {
	app := &Application{}
	req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()
	app.handleServiceValidate().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleServiceValidate_InvalidConfig(t *testing.T) {
	app := &Application{}
	// Missing name
	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://example.com")},
		},
	}
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleServiceValidate().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, false, resp["valid"])
	assert.Contains(t, resp["details"], "Static validation failed")
}

func TestHandleServiceValidate_ConnectivityFailure(t *testing.T) {
	// Need to allow private IPs for this test as we are connecting to localhost
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	app := &Application{}
	// Use a port that is likely closed
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://127.0.0.1:45678")},
		},
	}
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleServiceValidate().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // It returns 200 OK even on failure, but valid=false
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, false, resp["valid"])
	assert.Contains(t, resp["details"], "Connectivity check failed")
	// The error message depends on OS/network stack, usually "connection refused"
	assert.NotNil(t, resp["error"])
}

func TestHandleServiceValidate_Success(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	app := &Application{}
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{Address: proto.String(ts.URL)},
		},
	}
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleServiceValidate().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, true, resp["valid"])
}
