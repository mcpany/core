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

func TestUpstreamServiceManager_LoadAndMergeCollection_GitHub(t *testing.T) {
	// Mock GitHub API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("DEBUG: Requested %s\n", r.URL.Path)
		if r.URL.Path == "/repos/owner/repo/contents/path" {
			// List directory
			w.Header().Set("Content-Type", "application/json")
			// Return a file and a directory
			_, _ = fmt.Fprintln(w, `[
				{
					"type": "file",
					"name": "service.yaml",
					"html_url": "https://github.com/owner/repo/blob/main/path/service.yaml",
					"download_url": "http://`+r.Host+`/raw/service.yaml"
				},
				{
					"type": "dir",
					"name": "subdir",
					"html_url": "https://github.com/owner/repo/tree/main/path/subdir",
					"url": "http://`+r.Host+`/repos/owner/repo/contents/path/subdir"
				}
			]`)
			return
		}
		if r.URL.Path == "/raw/service.yaml" {
			// Return config file content
			_, _ = fmt.Fprint(w, `
services:
  - name: github-service-1
    http_service:
      address: http://example.com
`)
			return
		}

		// Handle subdir list (recursive check)
		// loadAndMergeCollection calls newCollection for subdir
		// The URL for subdir will be https://github.com/owner/repo/tree/main/path/subdir (from html_url)
		// We need to mock the LIST call for that too if we want to test recursion fully.
		// For now, let's just error on subdir to test error handling or make it empty.
		if r.URL.Path == "/repos/owner/repo/contents/path/subdir" {
			_, _ = fmt.Fprintln(w, `[]`)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	m := NewUpstreamServiceManager(nil)
	// Override httpClient to allow 127.0.0.1 requests (bypass SSRF protection)
	m.httpClient = ts.Client()

	// Override newGitHub
	m.newGitHub = func(_ context.Context, rawURL string) (*GitHub, error) {
		// Simple manual parsing or use the regex from validator/github.go if accessible.
		// But since we control the input URLs, we can just split.
		// Input: https://github.com/owner/repo/tree/main/path
		// or: https://github.com/owner/repo/tree/main/path/subdir

		path := "path"
		if rawURL == "https://github.com/owner/repo/tree/main/path/subdir" {
			path = "path/subdir"
		}

		return &GitHub{
			Owner:         "owner",
			Repo:          "repo",
			Path:          path,
			Ref:           "main",
			URLType:       "tree",
			log:           m.log,
			apiURL:        ts.URL, // Point to test server
			rawContentURL: ts.URL, // Point to test server for downloads
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
	assert.Len(t, m.services, 1)
	assert.Contains(t, m.services, "github-service-1")
}

func TestUpstreamServiceManager_LoadAndMergeCollection_GitHub_Error(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	m.newGitHub = func(_ context.Context, _ string) (*GitHub, error) {
		return nil, fmt.Errorf("mock setup error")
	}

	collection := configv1.Collection_builder{
		Name:    proto.String("test-collection"),
		HttpUrl: proto.String("https://github.com/owner/repo/tree/main/path"),
	}.Build()

	err := m.loadAndMergeCollection(context.Background(), collection)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse github url")
	assert.Contains(t, err.Error(), "mock setup error")
}

func TestUpstreamServiceManager_LoadAndMergeCollection_GitHub_ListError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "api error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	m := NewUpstreamServiceManager(nil)
	m.newGitHub = func(_ context.Context, _ string) (*GitHub, error) {
		return &GitHub{
			Owner:      "owner",
			Repo:       "repo",
			Path:       "path",
			apiURL:     ts.URL,
			httpClient: ts.Client(),
		}, nil
	}

	collection := configv1.Collection_builder{
		HttpUrl: proto.String("https://github.com/owner/repo/tree/main/path"),
	}.Build()

	err := m.loadAndMergeCollection(context.Background(), collection)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list github directory")
}
