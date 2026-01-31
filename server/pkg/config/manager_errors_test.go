// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestUpstreamServiceManager_LoadAndMergeServices_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.URL.Path == "/bad-yaml" {
			_, _ = w.Write([]byte(`bad yaml content:`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Run("HTTP Non-200 Error", func(t *testing.T) {
		config := func() *configv1.McpAnyServerConfig {
			cfg := &configv1.McpAnyServerConfig{}
			col := &configv1.Collection{}
			col.SetName("error-collection")
			col.SetHttpUrl(server.URL + "/error")
			cfg.SetCollections([]*configv1.Collection{col})
			return cfg
		}()
		manager := NewUpstreamServiceManager(nil)
		manager.httpClient = &http.Client{}

		_, err := manager.LoadAndMergeServices(context.Background(), config)
		assert.NoError(t, err)
	})
}
