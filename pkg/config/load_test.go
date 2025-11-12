
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary config file
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config.textproto")
	err := os.WriteFile(filePath, []byte(content), 0o644)
	require.NoError(t, err, "Failed to write temp config file")
	return filePath
}

// TestLoadServices_ValidConfigs tests loading of various valid service configurations.
func TestLoadServices_ValidConfigs(t *testing.T) {
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
		address: "grpc://localhost:50051"
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
				assert.Equal(t, "grpc://localhost:50051", grpcService.GetAddress())
				assert.True(t, grpcService.GetUseReflection())
			},
		},
		{
			name: "valid http service with api key auth",
			textprotoContent: `
upstream_services: {
	name: "http-svc-1"
	upstream_authentication: {
		api_key: {
			header_name: "X-Token"
			api_key: { plain_text: "secretapikey" }
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
				assert.Equal(t, "http-svc-1", s.GetName())
				httpService := s.GetHttpService()
				require.NotNil(t, httpService)
				assert.Equal(t, "http://api.example.com/v1", httpService.GetAddress())
				auth := s.GetUpstreamAuthentication()
				require.NotNil(t, auth)
				apiKey := auth.GetApiKey()
				require.NotNil(t, apiKey)
				assert.Equal(t, "X-Token", apiKey.GetHeaderName())
				apiKeyValue, err := util.ResolveSecret(apiKey.GetApiKey())
				require.NoError(t, err)
				assert.Equal(t, "secretapikey", apiKeyValue, "API key should be plaintext")
			},
		},
		{
			name: "valid http service with bearer token auth",
			textprotoContent: `
upstream_services: {
	name: "http-svc-bearer"
	upstream_authentication: {
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
				auth := s.GetUpstreamAuthentication()
				require.NotNil(t, auth)
				bearerToken := auth.GetBearerToken()
				require.NotNil(t, bearerToken)
				tokenValue, err := util.ResolveSecret(bearerToken.GetToken())
				require.NoError(t, err)
				assert.Equal(t, "secretbearertoken", tokenValue)
			},
		},
		{
			name: "valid http service with basic auth",
			textprotoContent: `
upstream_services: {
	name: "http-svc-basic"
	upstream_authentication: {
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
				auth := s.GetUpstreamAuthentication()
				require.NotNil(t, auth)
				basicAuth := auth.GetBasicAuth()
				require.NotNil(t, basicAuth)
				assert.Equal(t, "testuser", basicAuth.GetUsername())
				passwordValue, err := util.ResolveSecret(basicAuth.GetPassword())
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
			expectedCount: 1,
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				assert.Equal(t, "duplicate-name", s.GetName())
				httpService := s.GetHttpService()
				require.NotNil(t, httpService)
				assert.Equal(t, "http://api.example.com/v1", httpService.GetAddress())
			},
		},
		{
			name: "detailed error for duplicate service names",
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
			expectedCount: 1,
			checkServices: func(t *testing.T, services []*configv1.UpstreamServiceConfig) {
				s := services[0]
				assert.Equal(t, "duplicate-name", s.GetName())
				httpService := s.GetHttpService()
				require.NotNil(t, httpService)
				assert.Equal(t, "http://api.example.com/v1", httpService.GetAddress())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempConfigFile(t, tt.textprotoContent)
			fs := afero.NewOsFs()
			fileStore := NewFileStore(fs, []string{filePath})
			cfg, err := LoadServices(fileStore, "server")

			if tt.expectLoadError {
				assert.Error(t, err)
				if tt.name == "detailed error for duplicate service names" {
					assert.Contains(t, err.Error(), "service 'duplicate-name': duplicate service name found")
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
