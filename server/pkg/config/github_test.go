package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/proto"
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
		{
			name:          "Valid GitHub URL without tree/blob",
			url:           "https://github.com/mcpany/core",
			expectedError: false,
			expected: &GitHub{
				Owner:   "mcpany",
				Repo:    "core",
				Ref:     "main",
				Path:    "",
				URLType: "tree",
			},
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
		Owner:         "mcpany",
		Repo:          "core",
		Ref:           "main",
		Path:          "README.md",
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
		_, _ = w.Write([]byte(`[{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/README.md", "download_url": "https://raw.githubusercontent.com/mcpany/core/main/examples/README.md"}]`))
	}))
	defer server.Close()

	g := &GitHub{
		Owner:      "mcpany",
		Repo:       "core",
		Ref:        "main",
		Path:       "examples",
		apiURL:     server.URL,
		httpClient: &http.Client{},
	}

	auth := configv1.Authentication_builder{
		BearerToken: configv1.BearerTokenAuth_builder{
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("my-secret-token"),
			}.Build(),
		}.Build(),
	}.Build()

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
	if len(contents) != 1 {
		t.Errorf("List() returned %d items, want 1. Contents: %v", len(contents), contents)
	} else if contents[0] != expected[0] {
		t.Errorf("List() = %v, want %v", contents, expected)
	}
}

func TestGitHub_List_With_Single_File(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"type": "file", "html_url": "https://github.com/mcpany/core/blob/main/examples/README.md", "download_url": "https://raw.githubusercontent.com/mcpany/core/main/examples/README.md"}`))
	}))
	defer server.Close()

	g := &GitHub{
		Owner:      "mcpany",
		Repo:       "core",
		Ref:        "main",
		Path:       "examples",
		apiURL:     server.URL,
		httpClient: &http.Client{},
	}

	contents, err := g.List(context.Background(), nil)
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
	if len(contents) != 1 {
		t.Errorf("List() returned %d items, want 1. Contents: %v", len(contents), contents)
	} else if contents[0] != expected[0] {
		t.Errorf("List() = %v, want %v", contents, expected)
	}
}

func TestGitHub_List_ssrf(t *testing.T) {
	// This test verifies that the GitHub client blocks requests to loopback addresses.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"type": "file"}]`))
	}))
	defer server.Close()

	// The httptest.Server URL uses a loopback address (e.g., 127.0.0.1), which should be blocked.
	g := &GitHub{
		Owner:  "mcpany",
		Repo:   "core",
		Ref:    "main",
		Path:   "examples",
		apiURL: server.URL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: util.SafeDialContext,
			},
		},
	}

	_, err := g.List(context.Background(), nil)
	if err == nil {
		t.Errorf("Expected an error due to SSRF attempt, but got nil")
	} else if !strings.Contains(err.Error(), "ssrf attempt blocked") {
		t.Errorf("Expected error to contain 'ssrf attempt blocked', but got: %v", err)
	}
}

func TestGitHub_List_Auth_Variants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API Key
		if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
			if apiKey != "my-api-key" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
			return
		}

		// Check Basic Auth
		if user, pass, ok := r.BasicAuth(); ok {
			if user != "user" || pass != "pass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	g := &GitHub{
		Owner:      "mcpany",
		Repo:       "core",
		Ref:        "main",
		Path:       "examples",
		apiURL:     server.URL,
		httpClient: &http.Client{},
	}

	ctx := context.Background()

	t.Run("APIKey", func(t *testing.T) {
		auth := configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value: configv1.SecretValue_builder{
					PlainText: proto.String("my-api-key"),
				}.Build(),
			}.Build(),
		}.Build()

		_, err := g.List(ctx, auth)
		if err != nil {
			t.Errorf("List() with API Key error = %v", err)
		}
	})

	t.Run("BasicAuth", func(t *testing.T) {
		auth := configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username: proto.String("user"),
				Password: configv1.SecretValue_builder{
					PlainText: proto.String("pass"),
				}.Build(),
			}.Build(),
		}.Build()

		_, err := g.List(ctx, auth)
		if err != nil {
			t.Errorf("List() with Basic Auth error = %v", err)
		}
	})
}
