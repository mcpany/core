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
			var res ResourceMCPServer
			if err := json.NewDecoder(r.Body).Decode(&res); err != nil || res.Name == "fail" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/api/v1/servers/test" {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(ResourceMCPServer{Name: "test", Port: 9090, Enabled: true}); err != nil {
				return
			}
			return
		}
		if r.Method == "GET" && r.URL.Path == "/api/v1/servers/error" {
			w.WriteHeader(http.StatusInternalServerError)
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

	// Test Read NotFound.
	resNotFound, err := Read(context.Background(), ts.URL, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, resNotFound)

	// Test Create Failure.
	err = Create(context.Background(), ts.URL, &ResourceMCPServer{Name: "fail", Port: 9090})
	assert.Error(t, err)

	// Test Read Error.
	resErr, err := Read(context.Background(), ts.URL, "error")
	assert.Error(t, err)
	assert.Nil(t, resErr)

	// Test Invalid URL.
	_, err = Read(context.Background(), "http://\x00invalid", "test")
	assert.Error(t, err)
	err = Create(context.Background(), "http://\x00invalid", &ResourceMCPServer{Name: "test", Port: 9090})
	assert.Error(t, err)

}
