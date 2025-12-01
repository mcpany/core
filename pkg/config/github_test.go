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
	"net/http"
	"net/http/httptest"
	"testing"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestIsGitHubURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Valid GitHub URL",
			url:      "https://github.com/mcpany/core/blob/main/README.md",
			expected: true,
		},
		{
			name:     "Valid GitHub Directory URL",
			url:      "https://github.com/mcpany/core/tree/main/examples",
			expected: true,
		},
		{
			name:     "Invalid GitHub URL",
			url:      "https://example.com",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isGitHubURL(tc.url); got != tc.expected {
				t.Errorf("isGitHubURL(%q) = %v, want %v", tc.url, got, tc.expected)
			}
		})
	}
}

func TestNewGitHub(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		expectedError bool
		expected      *GitHub
	}{
		{
			name:          "Valid GitHub URL",
			url:           "https://github.com/mcpany/core/blob/main/README.md",
			expectedError: false,
			expected: &GitHub{
				Owner:   "mcpany",
				Repo:    "core",
				Ref:     "main",
				Path:    "README.md",
				URLType: "blob",
			},
		},
		{
			name:          "Valid GitHub Directory URL",
			url:           "https://github.com/mcpany/core/tree/main/examples",
			expectedError: false,
			expected: &GitHub{
				Owner:   "mcpany",
				Repo:    "core",
				Ref:     "main",
				Path:    "examples",
				URLType: "tree",
			},
		},
		{
			name:          "Invalid GitHub URL",
			url:           "https://example.com",
			expectedError: true,
			expected:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGitHub(context.Background(), tc.url)
			if (err != nil) != tc.expectedError {
				t.Fatalf("NewGitHub() error = %v, expectedError %v", err, tc.expectedError)
			}
			if !tc.expectedError {
				if g.Owner != tc.expected.Owner || g.Repo != tc.expected.Repo || g.Ref != tc.expected.Ref || g.Path != tc.expected.Path || g.URLType != tc.expected.URLType {
					t.Errorf("NewGitHub() = %+v, want %+v", g, tc.expected)
				}
			}
		})
	}
}

func TestGitHub_ToRawContentURL(t *testing.T) {
	g := &GitHub{
		Owner:       "mcpany",
		Repo:        "core",
		Ref:         "main",
		Path:        "README.md",
		rawContentURL: "https://raw.githubusercontent.com",
	}
	expected := "https://raw.githubusercontent.com/mcpany/core/main/README.md"
	if got := g.ToRawContentURL(); got != expected {
		t.Errorf("ToRawContentURL() = %q, want %q", got, expected)
	}
}

func TestGitHub_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer my-secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/README.md", "download_url": "https://raw.githubusercontent.com/mcpany/core/main/examples/README.md"}]`))
	}))
	defer server.Close()

	g := &GitHub{
		Owner:   "mcpany",
		Repo:    "core",
		Ref:     "main",
		Path:    "examples",
		apiURL:  server.URL,
	}

	auth := &configv1.UpstreamAuthentication{}
	secret := &configv1.SecretValue{}
	secret.SetPlainText("my-secret-token")
	bearer := &configv1.UpstreamBearerTokenAuth{}
	bearer.SetToken(secret)
	auth.SetBearerToken(bearer)

	contents, err := g.List(context.Background(), auth)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	expected := []Content{
		{
			Type:        "file",
			HTMLURL:     "https://github.com/mcpany/core/blob/main/examples/README.md",
			DownloadURL: "https://raw.githubusercontent.com/mcpany/core/main/examples/README.md",
		},
	}
	if len(contents) != 1 || contents[0] != expected[0] {
		t.Errorf("List() = %v, want %v", contents, expected)
	}
}
