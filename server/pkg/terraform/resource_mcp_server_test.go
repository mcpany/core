// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraformResource(t *testing.T) {
	schema := Schema()
	assert.Contains(t, schema, "name")
	assert.Contains(t, schema, "port")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/servers" {
			w.WriteHeader(http.StatusCreated)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/api/v1/servers/test" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ResourceMCPServer{Name: "test", Port: 9090, Enabled: true}) //nolint:errcheck
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	err := Create(context.Background(), ts.URL, &ResourceMCPServer{Name: "test", Port: 9090})
	assert.NoError(t, err)

	res, err := Read(context.Background(), ts.URL, "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", res.Name)
}
