// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamServiceManager_LoadAndMergeCollection_GitHub_CaseInsensitive(t *testing.T) {
	// Mock GitHub API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/path" {
			// List directory
			w.Header().Set("Content-Type", "application/json")
			// Return a file with uppercase extension
			_, _ = fmt.Fprintln(w, `[
				{
					"type": "file",
					"name": "service.YAML",
					"html_url": "https://github.com/owner/repo/blob/main/path/service.YAML",
					"download_url": "http://`+r.Host+`/raw/service.YAML"
				}
			]`)
			return
		}
		if r.URL.Path == "/raw/service.YAML" {
			// Return config file content
			_, _ = fmt.Fprint(w, `
services:
  - name: github-service-uppercase
    http_service:
      address: http://example.com
`)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	m := NewUpstreamServiceManager(nil)
	m.httpClient = ts.Client()

	m.newGitHub = func(_ context.Context, _ string) (*GitHub, error) {
		return &GitHub{
			Owner:         "owner",
			Repo:          "repo",
			Path:          "path",
			Ref:           "main",
			URLType:       "tree",
			log:           m.log,
			apiURL:        ts.URL,
			rawContentURL: ts.URL,
			httpClient:    ts.Client(),
		}, nil
	}

	collection := configv1.Collection_builder{
		Name:    proto.String("test-collection"),
		HttpUrl: proto.String("https://github.com/owner/repo/tree/main/path"),
	}.Build()

	err := m.loadAndMergeCollection(context.Background(), collection)
	require.NoError(t, err)

	// Verify service loaded
	// This should FAIL if the extension check is case-sensitive
	assert.Len(t, m.services, 1, "Service from uppercase .YAML file should be loaded")
	assert.Contains(t, m.services, "github-service-uppercase")
}
