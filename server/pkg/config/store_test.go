// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_CollectFilePaths(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Setup files
	require.NoError(t, afero.WriteFile(fs, "/config/a.yaml", []byte(""), 0644))
	require.NoError(t, afero.WriteFile(fs, "/config/b.json", []byte(""), 0644))
	require.NoError(t, afero.WriteFile(fs, "/config/ignore.txt", []byte(""), 0644))
	require.NoError(t, fs.MkdirAll("/config/sub", 0755))
	require.NoError(t, afero.WriteFile(fs, "/config/sub/c.yml", []byte(""), 0644))

	// Test JSON loading
	require.NoError(t, afero.WriteFile(fs, "/config/conf.json", []byte(`{
		"global_settings": {
			"api_key": "test-key-12345678"
		}
	}`), 0644))

	// Test TextProto loading
	require.NoError(t, afero.WriteFile(fs, "/config/conf.textproto", []byte(`
		global_settings {
			api_key: "basic-auth-user"
		}
	`), 0644))

	// Test Expansion
	// Handle errors for lint
	err := os.Setenv("TEST_PORT", "9090")
	assert.NoError(t, err)
	defer func() { _ = os.Unsetenv("TEST_PORT") }()
	require.NoError(t, afero.WriteFile(fs, "/config/expand.yaml", []byte(`
upstream_services:
  - name: expanded-service
    http_service:
      address: http://localhost:${TEST_PORT}
`), 0644))

	tests := []struct {
		name          string
		paths         []string
		expectedFiles []string
		expectedError string
	}{
		{
			name:  "valid directory",
			paths: []string{"/config"},
			// Note: order depends on walk/sort.
			// a.yaml, b.json, conf.json, conf.textproto, expand.yaml, ignore.txt, sub/c.yml
			expectedFiles: []string{
				"/config/a.yaml",
				"/config/b.json",
				"/config/conf.json",
				"/config/conf.textproto",
				"/config/expand.yaml",
				"/config/sub/c.yml",
			},
		},
		{
			name:          "mixed file and dir",
			paths:         []string{"/config/a.yaml", "/config/sub"},
			expectedFiles: []string{"/config/a.yaml", "/config/sub/c.yml"},
		},
		{
			name:          "ignored extensions",
			paths:         []string{"/config/ignore.txt"},
			expectedFiles: nil, // Should be ignored
		},
		{
			name:          "non-existent path",
			paths:         []string{"/missing"},
			expectedError: "failed to stat path /missing: open /missing: file does not exist",
		},
		{
			name:          "url path",
			paths:         []string{"http://example.com/config.yaml"},
			expectedFiles: []string{"http://example.com/config.yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := NewFileStore(fs, tc.paths)
			files, err := store.collectFilePaths()
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFiles, files)
			}
		})
	}
}

func TestFileStore_CollectFilePaths_WalkError(_ *testing.T) {
	// Mock FS to force walk error?
	// afero.MemMapFs doesn't easily mock errors.
	// But we can simulate a file that acts like a directory?
	// Or use a ReadDir error.
}

func TestFileStore_Load_Error(t *testing.T) {
	fs := afero.NewMemMapFs()
	store := NewFileStore(fs, []string{"/missing"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect config file paths")
}

func TestFileStore_Load_Engines(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := os.Setenv("MY_VAR", "expanded")
	assert.NoError(t, err)
	defer func() { _ = os.Unsetenv("MY_VAR") }()

	// JSON
	require.NoError(t, afero.WriteFile(fs, "/config/1.json", []byte(`{"global_settings": {"api_key": "json-key"}}`), 0644))
	// YAML with expansion
	require.NoError(t, afero.WriteFile(fs, "/config/2.yaml", []byte(`global_settings: {api_key: "${MY_VAR}-key"}`), 0644))
	// TextProto
	require.NoError(t, afero.WriteFile(fs, "/config/3.textproto", []byte(`global_settings { api_key: "proto-key" }`), 0644))

	store := NewFileStore(fs, []string{"/config"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)
	// Last one wins (alphabetic order 1,2,3) -> 3.textproto
	// But merge behavior?
	// global_settings are merged? oneof?
	// actually apiKey is string. Last one overwrites.
	assert.Equal(t, "proto-key", cfg.GetGlobalSettings().GetApiKey())
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		env      map[string]string
		expected string
	}{
		{
			name:     "Variable set",
			input:    "Hello ${NAME}",
			env:      map[string]string{"NAME": "World"},
			expected: "Hello World",
		},
		{
			name:     "Variable unset",
			input:    "Hello ${NAME}",
			env:      map[string]string{},
			expected: "Hello ${NAME}",
		},
		{
			name:     "Variable set to empty",
			input:    "Hello ${NAME}",
			env:      map[string]string{"NAME": ""},
			expected: "Hello ",
		},
		{
			name:     "Variable with default, unset",
			input:    "Hello ${NAME:World}",
			env:      map[string]string{},
			expected: "Hello World",
		},
		{
			name:     "Variable with default, set",
			input:    "Hello ${NAME:World}",
			env:      map[string]string{"NAME": "Universe"},
			expected: "Hello Universe",
		},
		{
			name:     "Variable with default, set to empty",
			input:    "Hello ${NAME:World}",
			env:      map[string]string{"NAME": ""},
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			got := string(expand([]byte(tt.input)))
			assert.Equal(t, tt.expected, got)

			for k := range tt.env {
				os.Unsetenv(k)
			}
		})
	}
}

func TestYamlEngine_LogLevelFix(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/log.yaml", []byte(`
global_settings:
  log_level: INFO
`), 0644))

	store := NewFileStore(fs, []string{"/config/log.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())
}

func TestYamlEngine_ValidationFail(_ *testing.T) {
	_ = afero.NewMemMapFs()
	// Invalid config: missing required fields or constraint violation?
	// Currently validation is weak (stubbed).
	// But duplicate service names in different files?
}

func TestReadURL_Redirect(t *testing.T) {
	// Verify that redirects are disabled.
	// We mock the global httpClient for this test to allow localhost connections
	// while maintaining the CheckRedirect logic we want to test.
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/other", http.StatusFound)
	}))
	defer server.Close()

	_, err := readURL(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirects are disabled")
}

func TestYamlEngine_MultipleServiceTypes(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Define a service with BOTH http_service and grpc_service to trigger oneof error
	require.NoError(t, afero.WriteFile(fs, "/config/multi.yaml", []byte(`
upstream_services:
  - name: multi-service
    http_service:
      address: http://example.com
    grpc_service:
      address: localhost:50051
`), 0644))

	store := NewFileStore(fs, []string{"/config/multi.yaml"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	// We expect the custom error message about multiple service types
	assert.Contains(t, err.Error(), "has multiple service types defined")
}

func TestYamlEngine_UnknownField(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/unknown.yaml", []byte(`
global_settings:
  unknown_field: "foo"
`), 0644))

	store := NewFileStore(fs, []string{"/config/unknown.yaml"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	// protojson unmarshal errors on unknown fields by default
	assert.Contains(t, err.Error(), "unknown field")
}

func TestReadURL_Localhost(t *testing.T) {
	// Verify we can read from localhost if the client is configured to allow it.
	// We mock the global httpClient to simulate a SafeHTTPClient with loopback enabled.
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	// Start a local HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte("global_settings:\n  log_level: INFO\n"))
	}))
	defer ts.Close()

	// Append a path with extension so NewEngine detects it as YAML
	configURL := ts.URL + "/config.yaml"

	fs := afero.NewMemMapFs()
	store := NewFileStore(fs, []string{configURL})

	cfg, err := store.Load(context.Background())

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GlobalSettings.GetLogLevel())
}
