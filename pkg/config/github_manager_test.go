// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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
)

func TestUpstreamServiceManager_LoadAndMergeServices_GitHub(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/mcpany/core/contents/examples":
			w.Write([]byte(fmt.Sprintf(`[
				{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/service.yaml", "download_url": "%s/service.yaml"},
				{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/README.md", "download_url": "%s/README.md"},
				{"type": "dir", "html_url": "https://github.com/mcpany/core/tree/main/examples/nested", "download_url": null}
			]`, server.URL, server.URL)))
		case "/service.yaml":
			w.Write([]byte(`{"services": [{"name": "github-service", "version": "1.0"}]}`))
		case "/repos/mcpany/core/contents/examples/nested":
			w.Write([]byte(fmt.Sprintf(`[
				{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/nested/service.json", "download_url": "%s/service.json"}
			]`, server.URL)))
		case "/service.json":
			w.Write([]byte(`{"services": [{"name": "nested-service", "version": "1.0"}]}`))
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
	collection := &configv1.UpstreamServiceCollection{}
	collection.SetName("github-dir")
	collection.SetHttpUrl("https://github.com/mcpany/core/tree/main/examples")
	auth := &configv1.UpstreamAuthentication{}
	secret := &configv1.SecretValue{}
	secret.SetPlainText("my-secret-token")
	bearer := &configv1.UpstreamBearerTokenAuth{}
	bearer.SetToken(secret)
	auth.SetBearerToken(bearer)
	collection.SetAuthentication(auth)
	config := &configv1.McpAnyServerConfig{}
	config.SetUpstreamServiceCollections([]*configv1.UpstreamServiceCollection{collection})

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
