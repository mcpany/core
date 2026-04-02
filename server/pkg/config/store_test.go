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
      address: http://127.0.0.1:${TEST_PORT}
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

type mockErrorFs struct {
	afero.Fs
	errToReturn error
}

func (m *mockErrorFs) Stat(name string) (os.FileInfo, error) {
	return nil, m.errToReturn
}

func (m *mockErrorFs) Open(name string) (afero.File, error) {
	return nil, m.errToReturn
}

func TestFileStore_CollectFilePaths_WalkError(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/file.yaml", []byte(""), 0644))

	mockFs := &mockErrorFs{Fs: fs, errToReturn: os.ErrPermission}

	store := NewFileStore(mockFs, []string{"/config"})
	_, err := store.collectFilePaths()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestFileStore_Load_Error(t *testing.T) {
	fs := afero.NewMemMapFs()
	store := NewFileStore(fs, []string{"/missing"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect config file paths")
}

func TestFileStore_LoadOneConfig_Errors(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Add an unreadable file by wrapping fs in a mock that fails ReadFile
	mockFs := &mockErrorFs{Fs: fs, errToReturn: os.ErrPermission}
	store := NewFileStore(mockFs, []string{"/config/file.yaml"})

	// Create the file so collect doesn't fail but read fails
	require.NoError(t, afero.WriteFile(fs, "/config/file.yaml", []byte(""), 0644))

	_, err := store.loadOneConfig(context.Background(), "/config/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")

	// Engine error when loading a bad extension
	require.NoError(t, afero.WriteFile(fs, "/config/file.badext", []byte("content"), 0644))

	store = NewFileStore(fs, []string{"/config/file.badext"})
	_, err = store.loadOneConfig(context.Background(), "/config/file.badext")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config file extension")

	// Same error but with skipErrors
	storeWithSkip := NewFileStoreWithSkipErrors(fs, []string{"/config/file.badext"})
	cfg, errSkip := storeWithSkip.loadOneConfig(context.Background(), "/config/file.badext")
	assert.NoError(t, errSkip) // error should be suppressed
	assert.Nil(t, cfg)
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
		name        string
		input       string
		env         map[string]string
		expected    string
		expectError bool
	}{
		{
			name:        "Variable set",
			input:       "Hello ${NAME}",
			env:         map[string]string{"NAME": "World"},
			expected:    "Hello World",
			expectError: false,
		},
		{
			name:        "Variable unset",
			input:       "Hello ${NAME}",
			env:         map[string]string{},
			expected:    "",
			expectError: true,
		},
		{
			name:        "Variable set to empty",
			input:       "Hello ${NAME}",
			env:         map[string]string{"NAME": ""},
			expected:    "Hello ",
			expectError: false,
		},
		{
			name:        "Variable with default, unset",
			input:       "Hello ${NAME:World}",
			env:         map[string]string{},
			expected:    "Hello World",
			expectError: false,
		},
		{
			name:        "Variable with default, set",
			input:       "Hello ${NAME:World}",
			env:         map[string]string{"NAME": "Universe"},
			expected:    "Hello Universe",
			expectError: false,
		},
		{
			name:        "Variable with default, set to empty",
			input:       "Hello ${NAME:World}",
			env:         map[string]string{"NAME": ""},
			expected:    "Hello World",
			expectError: false,
		},
		{
			name:        "Short syntax variable",
			input:       "Hello $NAME",
			env:         map[string]string{"NAME": "World"},
			expected:    "Hello World",
			expectError: false,
		},
		{
			name:        "Short syntax variable with text following",
			input:       "Hello $NAME!",
			env:         map[string]string{"NAME": "World"},
			expected:    "Hello World!",
			expectError: false,
		},
		{
			name:        "Short syntax variable with underscore",
			input:       "Hello $MY_NAME",
			env:         map[string]string{"MY_NAME": "World"},
			expected:    "Hello World",
			expectError: false,
		},
		{
			name:        "Short syntax variable missing (no default supported)",
			input:       "Hello $MISSING",
			env:         map[string]string{},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			}()

			got, err := expand([]byte(tt.input))
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(got))
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

func TestApplyPathToMap(t *testing.T) {
	// 1. "upstream" should be rewritten to "upstream_services"
	m := make(map[string]interface{})
	var cfg configv1.McpAnyServerConfig
	applyPathToMap(m, []string{"upstream"}, "foo", &cfg)
	// Because "upstream_services" is a list field, applyPathToMap converts a direct value into a list.
	// So we expect a slice.
	assert.Equal(t, []interface{}{"foo"}, m["upstream_services"])

	// 2. Deep path setting on empty map
	m2 := make(map[string]interface{})
	applyPathToMap(m2, []string{"global_settings", "log_level"}, "DEBUG", &cfg)

	gs, ok := m2["global_settings"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "DEBUG", gs["log_level"])
}

func TestResolveEnvValue(t *testing.T) {
	var cfg configv1.McpAnyServerConfig

	// Test field not found fallback
	val := resolveEnvValue(&cfg, []string{"non_existent_field"}, "value")
	assert.Equal(t, "value", val)

	// Test bool array fallback (malformed bool returns string)
	// Experimental features might not be a bool array, let's use a known string or bool array field if any exist.
	// The core config might not have repeated bools. But we can test parsing string.
	val2 := resolveEnvValue(&cfg, []string{"global_settings", "disable_auth"}, "invalid_bool")
	assert.Equal(t, "invalid_bool", val2)

	// Test path continuing past a scalar (mismatch)
	// global_settings.log_level is a scalar enum. Let's try adding another path segment.
	val3 := resolveEnvValue(&cfg, []string{"global_settings", "log_level", "extra"}, "value")
	assert.Equal(t, "value", val3)
}

func TestYamlEngine_ValidationFail(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Invalid config: passing services instead of upstream_services should throw an error
	require.NoError(t, afero.WriteFile(fs, "/config/invalid.yaml", []byte(`
services:
  - name: "wrapper-test"
    http_service:
      address: "https://example.com"
`), 0644))

	store := NewFileStore(fs, []string{"/config/invalid.yaml"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Did you mean \"upstream_services\"?")
}

func TestYamlEngine_SkipValidation(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/schema.yaml", []byte(`
global_settings:
  log_level: INVALID_LOG_LEVEL
`), 0644))

	store := NewFileStore(fs, []string{"/config/schema.yaml"})
	// Without skipValidation, validation shouldn't fail explicitly yet due to how protojson handles enums,
	// but let's test that the flag propagates.
	store.SetSkipValidation(true)
	_, err := store.Load(context.Background())
	// protojson will still fail to unmarshal an invalid enum string, returning an error.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value for enum field")
}

func TestSplitByCommaIgnoringBraces(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a,\"b,c\",d", []string{"a", "\"b,c\"", "d"}}, // Comma inside quotes
		{"a,{b,c},d", []string{"a", "{b,c}", "d"}},    // Comma inside braces
		{"a,[b,c],d", []string{"a", "[b,c]", "d"}},    // Comma inside brackets
		{"a,b\\,c,d", []string{"a", "b\\,c", "d"}},    // Escaped comma (or just escape sequence)
		{"a,\"b", []string{"a", "\"b"}},               // Unclosed quote (ends string)
		{"a,{b", []string{"a", "{b"}},                 // Unclosed brace
		{"a,\\", []string{"a", "\\"}},                 // Trailing escape
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res := splitByCommaIgnoringBraces(tc.input)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestUnquoteCSV(t *testing.T) {
	assert.Equal(t, "abc", unquoteCSV("\"abc\""))
	assert.Equal(t, "abc", unquoteCSV("abc"))
	assert.Equal(t, "\"abc", unquoteCSV("\"abc")) // Unclosed
	assert.Equal(t, "abc\"", unquoteCSV("abc\"")) // Unstarted
	assert.Equal(t, "ab\"c", unquoteCSV("\"ab\"\"c\"")) // Double quotes inside
}

func TestFindKeyLine(t *testing.T) {
	yamlBytes := []byte(`
global_settings:
  log_level: DEBUG
upstream_services: []
`)
	line := findKeyLine(yamlBytes, "log_level")
	assert.Equal(t, 3, line)

	invalidYaml := []byte(`
global_settings:
  log_level: DEBUG
  - invalid
`)
	lineInvalid := findKeyLine(invalidYaml, "log_level")
	assert.Equal(t, 0, lineInvalid) // Unmarshal error
}

func TestReadURL_Redirect(t *testing.T) {
	// Verify that redirects are disabled.
	// We mock the global httpClient for this test to allow 127.0.0.1 connections
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
      address: 127.0.0.1:50051
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

func TestYamlEngine_MCPServersField(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config that looks like Claude Desktop config
	// Note: @ needs to be quoted in YAML if it starts a value, but here it's in a list
	// The error "found character that cannot start any token" suggests issues with @
	require.NoError(t, afero.WriteFile(fs, "/config/claude_config.yaml", []byte(`
mcpServers:
  filesystem:
    command: "npx"
    args:
      - "-y"
      - "@modelcontextprotocol/server-filesystem"
      - "/path/to/allowed/files"
`), 0644))

	store := NewFileStore(fs, []string{"/config/claude_config.yaml"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	// Expect a helpful error message guiding the user
	assert.Contains(t, err.Error(), "Did you mean \"upstream_services\"?")
	assert.Contains(t, err.Error(), "mcpServers")
}

func TestReadURL_Localhost(t *testing.T) {
	// Verify we can read from 127.0.0.1 if the client is configured to allow it.
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
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())
}

func TestYamlEngine_ServiceConfigWrapper(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config that incorrectly uses service_config wrapper
	require.NoError(t, afero.WriteFile(fs, "/config/service_config_wrapper.yaml", []byte(`
upstream_services:
  - name: "wrapper-test"
    service_config:
      http_service:
        address: "https://example.com"
`), 0644))

	store := NewFileStore(fs, []string{"/config/service_config_wrapper.yaml"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
	// Expect the new helpful error message
	assert.Contains(t, err.Error(), "without a 'service_config' wrapper")
}
