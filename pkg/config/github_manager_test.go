/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v58/github"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestLoadGitHubCollection_File(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := `
services:
  - name: "test-service"
    http_service:
      address: "http://localhost:8080"
`
		encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
		w.Write([]byte(`{"content": "` + encodedContent + `"}`))
	}))
	defer server.Close()

	manager := NewUpstreamServiceManager()
	collection := &configv1.UpstreamServiceCollection{
		Name: "test-collection",
		Source: &configv1.UpstreamServiceCollection_Github_{
			Github: &configv1.GitHubCollection{
				Owner: "owner",
				Repo:  "repo",
				Path:  "path",
			},
		},
	}

	// This is a bit of a hack, but it allows us to inject the mock server
	// without changing the function signature.
	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = originalClient }()

	err := manager.loadAndMergeCollection(context.Background(), collection)
	assert.NoError(t, err)

	var services []*configv1.UpstreamServiceConfig
	for _, service := range manager.services {
		services = append(services, service)
	}

	assert.Len(t, services, 1)
	assert.Equal(t, "test-service", services[0].Name)
}

func TestLoadGitHubCollection_Directory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/repos/owner/repo/contents/path" {
			w.Write([]byte(`[{"type": "file", "path": "path/to/file1.yaml"}, {"type": "file", "path": "path/to/file2.yaml"}]`))
		} else if r.URL.Path == "/repos/owner/repo/contents/path/to/file1.yaml" {
			content := `
services:
  - name: "test-service-1"
    http_service:
      address: "http://localhost:8080"
`
			encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
			w.Write([]byte(`{"content": "` + encodedContent + `"}`))
		} else if r.URL.Path == "/repos/owner/repo/contents/path/to/file2.yaml" {
			content := `
services:
  - name: "test-service-2"
    http_service:
      address: "http://localhost:8080"
`
			encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
			w.Write([]byte(`{"content": "` + encodedContent + `"}`))
		}
	}))
	defer server.Close()

	manager := NewUpstreamServiceManager()
	collection := &configv1.UpstreamServiceCollection{
		Name: "test-collection",
		Source: &configv1.UpstreamServiceCollection_Github{
			Github: &configv1.GitHubCollection{
				Owner: "owner",
				Repo:  "repo",
				Path:  "path",
			},
		},
	}

	// This is a bit of a hack, but it allows us to inject the mock server
	// without changing the function signature.
	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = originalClient }()

	err := manager.loadAndMergeCollection(context.Background(), collection)
	assert.NoError(t, err)

	var services []*configv1.UpstreamServiceConfig
	for _, service := range manager.services {
		services = append(services, service)
	}

	assert.Len(t, services, 2)
	assert.Equal(t, "test-service-1", services[0].Name)
	assert.Equal(t, "test-service-2", services[1].Name)
}
