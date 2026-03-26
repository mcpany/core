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

func TestUpstreamServiceManager_LoadAndMergeServices_GitHub(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/mcpany/core/contents/examples":
			_, _ = fmt.Fprintf(w, `[
				{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/service.yaml", "download_url": "%s/service.yaml"},
				{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/README.md", "download_url": "%s/README.md"},
				{"type": "dir", "html_url": "https://github.com/mcpany/core/tree/main/examples/nested", "download_url": null}
			]`, server.URL, server.URL)
		case "/service.yaml":
			_, _ = w.Write([]byte(`{"services": [{"name": "github-service", "version": "1.0"}]}`))
		case "/repos/mcpany/core/contents/examples/nested":
			_, _ = fmt.Fprintf(w, `[
				{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/nested/service.json", "download_url": "%s/service.json"}
			]`, server.URL)
		case "/service.json":
			_, _ = w.Write([]byte(`{"services": [{"name": "nested-service", "version": "1.0"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalAPIURL := githubAPIURL
	originalRawContentURL := githubRawContentURL
	defer func() {
		githubAPIURL = originalAPIURL
		githubRawContentURL = originalRawContentURL
	}()

	githubAPIURL = server.URL
	githubRawContentURL = server.URL

	manager := NewUpstreamServiceManager(nil)
	manager.httpClient = &http.Client{}
	manager.newGitHub = func(ctx context.Context, rawURL string) (*GitHub, error) {
		g, err := NewGitHub(ctx, rawURL)
		if err != nil {
			return nil, err
		}
		g.httpClient = &http.Client{}
		return g, nil
	}
	collection := configv1.Collection_builder{
		Name:    proto.String("github-dir"),
		HttpUrl: proto.String("https://github.com/mcpany/core/tree/main/examples"),
		Authentication: configv1.Authentication_builder{
			BearerToken: configv1.BearerTokenAuth_builder{
				Token: configv1.SecretValue_builder{
					PlainText: proto.String("my-secret-token"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	config := configv1.McpAnyServerConfig_builder{
		Collections: []*configv1.Collection{collection},
	}.Build()

	loadedServices, err := manager.LoadAndMergeServices(context.Background(), config)
	require.NoError(t, err)

	expectedServiceNamesAndVersions := map[string]string{
		"github-service": "1.0",
		"nested-service": "1.0",
	}

	assert.Equal(t, len(expectedServiceNamesAndVersions), len(loadedServices))

	serviceMap := make(map[string]*configv1.UpstreamServiceConfig)
	for _, s := range loadedServices {
		serviceMap[s.GetName()] = s
	}

	for name, version := range expectedServiceNamesAndVersions {
		s, ok := serviceMap[name]
		assert.True(t, ok, "expected service %s to be loaded", name)
		assert.Equal(t, version, s.GetVersion())
	}
}
