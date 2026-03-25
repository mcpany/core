// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary config file
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config.textproto")
	err := os.WriteFile(filePath, []byte(content), 0o600)
	require.NoError(t, err, "Failed to write temp config file")
	return filePath
}

// TestLoadServices_ValidConfigs tests loading of various valid service configurations.
func TestLoadServices_ValidConfigs(t *testing.T) {
	t.Run("Load from URL", func(t *testing.T) {
		// Create a mock HTTP server to serve the config file
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`
upstream_services: {
	name: "http-svc-from-url"
	http_service: {
		address: "http://api.example.com/from-url"
	}
}
`))
		}))
		defer server.Close()
		originalClient := httpClient
		t.Cleanup(func() {
			httpClient = originalClient
		})
		httpClient = server.Client()

		fs := afero.NewOsFs()
		fileStore := NewFileStore(fs, []string{server.URL + "/config.textproto"})
		cfg, err := LoadServices(context.Background(), fileStore, "server")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Len(t, cfg.GetUpstreamServices(), 1)
		s := cfg.GetUpstreamServices()[0]
		assert.Equal(t, "http-svc-from-url", s.GetName())
		httpService := s.GetHttpService()
		require.NotNil(t, httpService)
		assert.Equal(t, "http://api.example.com/from-url", httpService.GetAddress())
	})

	t.Run("Load from URL with 404 error", func(t *testing.T) {
		// Create a mock HTTP server to serve the config file
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()
		originalClient := httpClient
		t.Cleanup(func() {
			httpClient = originalClient
		})
		httpClient = server.Client()

		fs := afero.NewOsFs()
		fileStore := NewFileStore(fs, []string{server.URL + "/config.textproto"})
		_, err := LoadServices(context.Background(), fileStore, "server")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status code 404")
	})

	t.Run("Load from URL with malformed content", func(t *testing.T) {
		// Create a mock HTTP server to serve the config file
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`
upstream_services: {
	name: "http-svc-from-url"
	http_service: {
		address: "http://api.example.com/from-url"
`))
		}))
		defer server.Close()
		originalClient := httpClient
		t.Cleanup(func() {
			httpClient = originalClient
		})
		httpClient = server.Client()

		fs := afero.NewOsFs()
		// Test with default NewFileStore (should error)
		fileStore := NewFileStore(fs, []string{server.URL + "/config.textproto"})
		_, err := LoadServices(context.Background(), fileStore, "server")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal config")

		// Test with NewFileStoreWithSkipErrors (should fail because empty config is now invalid if config sources are present)
		resilientStore := NewFileStoreWithSkipErrors(fs, []string{server.URL + "/config.textproto"})
		_, err = LoadServices(context.Background(), resilientStore, "server")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "configuration sources provided but loaded configuration is empty")
	})

	t.Run("Load from URL with response larger than 1MB", func(t *testing.T) {
		// Create a mock HTTP server to serve the config file
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1048577")
			_, _ = w.Write(make([]byte, 1048577))
		}))
		defer server.Close()
		originalClient := httpClient
		t.Cleanup(func() {
			httpClient = originalClient
		})
		httpClient = server.Client()

		fs := afero.NewOsFs()
		fileStore := NewFileStore(fs, []string{server.URL + "/config.textproto"})
		_, err := LoadServices(context.Background(), fileStore, "server")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request body too large")
	})

	t.Run("unknown binary type", func(t *testing.T) {
		// We need a valid config to reach the binary type check
		content := `
upstream_services: {
	name: "dummy-service"
	http_service: {
		address: "http://example.com"
	}
}
`
		filePath := createTempConfigFile(t, content)
		fs := afero.NewOsFs()
		fileStore := NewFileStore(fs, []string{filePath})
		_, err := LoadServices(context.Background(), fileStore, "unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown binary type: unknown")
	})
	tests := []struct {
		name             string
		textprotoContent string
		expectedCount    int
		checkServices    func(t *testing.T, services []*configv1.UpstreamServiceConfig)
		expectLoadError  bool
	}{
		{
			name: "valid grpc service with reflection",
			textprotoContent: `
upstream_services: {
	name: "grpc-svc-1"
	grpc_service: {
		address: "grpc://127.0.0.1:50051"
		use_reflection: true
	}
}
`,
			expectedCount: 1,
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				assert.Equal(t, "grpc-svc-1", s.GetName())
				grpcService := s.GetGrpcService()
				require.NotNil(t, grpcService)
				assert.Equal(t, "grpc://127.0.0.1:50051", grpcService.GetAddress())
				assert.True(t, grpcService.GetUseReflection())
			},
		},
		{
			name: "valid http service with api key auth",
			textprotoContent: `
upstream_services: {
	name: "http-svc-1"
	upstream_auth: {
		api_key: {
			param_name: "X-Token"
			value: { plain_text: "secretapikey" }
		}
	}
	http_service: {
		address: "http://api.example.com/v1"
		tools: {
			name: "get_user"
			call_id: "get_user_call"
		}
		calls: {
			key: "get_user_call"
			value: {
				id: "get_user_call"
			}
		}
		resources: {
			uri: "file:///test.txt"
			name: "test.txt"
			static: {
				text_content: "hello world"
			}
		}
	}
}
`,
			expectedCount: 1,
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				assert.Equal(t, "http-svc-1", s.GetName())
				httpService := s.GetHttpService()
				require.NotNil(t, httpService)
				assert.Equal(t, "http://api.example.com/v1", httpService.GetAddress())
				auth := s.GetUpstreamAuth()
				require.NotNil(t, auth)
				apiKey := auth.GetApiKey()
				require.NotNil(t, apiKey)
				assert.Equal(t, "X-Token", apiKey.GetParamName())
				apiKeyValue, err := util.ResolveSecret(context.Background(), apiKey.GetValue())
				require.NoError(t, err)
				assert.Equal(t, "secretapikey", apiKeyValue, "API key should be plaintext")
				assert.Len(t, httpService.GetTools(), 1)
				tool := httpService.GetTools()[0]
				assert.Equal(t, "get_user", tool.GetName())
				assert.Equal(t, "get_user_call", tool.GetCallId())
				assert.Contains(t, httpService.GetCalls(), "get_user_call")
				assert.Len(t, httpService.GetResources(), 1)
				assert.Equal(t, "file:///test.txt", httpService.GetResources()[0].GetUri())
			},
		},
		{
			name: "valid http service with bearer token auth",
			textprotoContent: `
upstream_services: {
	name: "http-svc-bearer"
	upstream_auth: {
		bearer_token: {
			token: { plain_text: "secretbearertoken" }
		}
	}
	http_service: {
		address: "http://api.example.com/v1"
	}
}
`,
			expectedCount: 1,
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				auth := s.GetUpstreamAuth()
				require.NotNil(t, auth)
				bearerToken := auth.GetBearerToken()
				require.NotNil(t, bearerToken)
				tokenValue, err := util.ResolveSecret(context.Background(), bearerToken.GetToken())
				require.NoError(t, err)
				assert.Equal(t, "secretbearertoken", tokenValue)
			},
		},
		{
			name: "valid http service with basic auth",
			textprotoContent: `
upstream_services: {
	name: "http-svc-basic"
	upstream_auth: {
		basic_auth: {
			username: "testuser"
			password: { plain_text: "secretpassword" }
		}
	}
	http_service: {
		address: "http://api.example.com/v1"
	}
}
`,
			expectedCount: 1,
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				auth := s.GetUpstreamAuth()
				require.NotNil(t, auth)
				basicAuth := auth.GetBasicAuth()
				require.NotNil(t, basicAuth)
				assert.Equal(t, "testuser", basicAuth.GetUsername())
				passwordValue, err := util.ResolveSecret(context.Background(), basicAuth.GetPassword())
				require.NoError(t, err)
				assert.Equal(t, "secretpassword", passwordValue)
			},
		},
		{
			name: "service with invalid cache config",
			textprotoContent: `
upstream_services: {
    name: "service-with-invalid-cache"
    http_service: {
        address: "http://api.example.com/v2"
    }
    cache: {
        is_enabled: true
        ttl: { seconds: -10 } # Invalid TTL
    }
}
`,
			expectLoadError: true,
			expectedCount:   0,
		},
		{
			name: "mixed valid and invalid services",
			textprotoContent: `
upstream_services: {
    name: "valid-service"
    http_service: {
        address: "http://valid.example.com"
    }
}
upstream_services: {
    name: "invalid-service"
    # Missing service type
}
`,
			expectLoadError: true,
			expectedCount:   0,
		},
		{
			name: "duplicate service names",
			textprotoContent: `
upstream_services: {
	name: "duplicate-name"
	http_service: {
		address: "http://api.example.com/v1"
	}
}
upstream_services: {
	name: "duplicate-name"
	http_service: {
		address: "http://api.example.com/v2"
	}
}
`,
			expectLoadError: false,
			expectedCount:   1, // Only the first one should be loaded
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				assert.Equal(t, "duplicate-name", s.GetName())
				httpService := s.GetHttpService()
				// The first one defined should be the one kept (assuming no priority override)
				// Wait, the manager logic says: "priority < existingPriority: New service has higher priority, replace".
				// "priority == existingPriority: Same priority, this is a duplicate" -> return nil (skip)
				// So the first one loaded is kept.
				assert.Equal(t, "http://api.example.com/v1", httpService.GetAddress())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempConfigFile(t, tt.textprotoContent)
			fs := afero.NewOsFs()
			fileStore := NewFileStore(fs, []string{filePath})
			cfg, err := LoadServices(context.Background(), fileStore, "server")

			if tt.expectLoadError {
				assert.Error(t, err)
				if tt.name == "duplicate service names" {
					assert.Contains(t, err.Error(), "duplicate service name found: duplicate-name")
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Len(t, cfg.GetUpstreamServices(), tt.expectedCount)
			if tt.checkServices != nil && len(cfg.GetUpstreamServices()) > 0 {
				tt.checkServices(t, cfg.GetUpstreamServices())
			} else if tt.expectedCount > 0 && len(cfg.GetUpstreamServices()) == 0 {
				t.Errorf("Expected %d services, but got 0", tt.expectedCount)
			}
		})
	}
}

func TestDefaultUserHasProfileAccessWhenIdIsMissing(t *testing.T) {
	content := `
global_settings: {
    profiles: "dev"
}
upstream_services: {
	name: "service-with-named-profile"
	http_service: {
		address: "http://api.example.com"
	}
}
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.textproto")
	fs := afero.NewOsFs()
	f, err := fs.Create(filePath)
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	fileStore := NewFileStore(fs, []string{filePath})
	cfg, err := LoadServices(context.Background(), fileStore, "server")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotEmpty(t, cfg.GetUsers())
	defaultUser := cfg.GetUsers()[0]
	assert.Equal(t, "default", defaultUser.GetId())

	// This should pass now with my fix
	assert.Contains(t, defaultUser.GetProfileIds(), "dev", "Default user should have access to 'dev' profile even if ID is not explicitly set")
}

func TestLoadServices_DefaultUser_ImplicitProfile(t *testing.T) {
	// Configuration with one service that has NO explicit profiles.
	// It should default to "default" profile.
	// No users defined, so a default user should be created.
	// The default user should have access to the "default" profile.
	content := `
upstream_services: {
	name: "service-implicit-profile"
	http_service: {
		address: "http://api.example.com"
	}
}
`
	filePath := createTempConfigFile(t, content)
	fs := afero.NewOsFs()
	fileStore := NewFileStore(fs, []string{filePath})

	// We use "server" binary type, and since we don't pass specific profiles via env/flags in this test helper directly,
	// GlobalSettings defaults might be used, but LoadServices internally initializes UpstreamServiceManager.
	// By default UpstreamServiceManager enables "default" profile if none provided.

	cfg, err := LoadServices(context.Background(), fileStore, "server")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check if default user was created
	require.Len(t, cfg.GetUsers(), 1)
	defaultUser := cfg.GetUsers()[0]
	assert.Equal(t, "default", defaultUser.GetId())

	// Check if default user has access to "default" profile
	assert.Contains(t, defaultUser.GetProfileIds(), "default", "Default user should have access to 'default' profile")
}

func TestDefaultUser_ShouldNotAccessDisabledProfiles(t *testing.T) {
	content := `
global_settings: {
    profiles: ["enabled_profile"]
    profile_definitions: [
        { name: "enabled_profile" },
        { name: "disabled_profile" }
    ]
}
`
	filePath := createTempConfigFile(t, content)
	fs := afero.NewOsFs()
	fileStore := NewFileStore(fs, []string{filePath})
	cfg, err := LoadServices(context.Background(), fileStore, "server")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotEmpty(t, cfg.GetUsers())
	defaultUser := cfg.GetUsers()[0]

	// Verify the default user has the enabled profile
	assert.Contains(t, defaultUser.GetProfileIds(), "enabled_profile")

	// Verify the default user DOES NOT have the disabled profile
	assert.NotContains(t, defaultUser.GetProfileIds(), "disabled_profile", "Default user should not have access to disabled profiles")
}
