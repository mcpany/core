// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func TestValidateSecretValue(t *testing.T) {
	tests := []struct {
		name      string
		secret    *configv1.SecretValue
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Nil secret",
			secret:    nil,
			expectErr: false,
		},
		{
			name: "Valid file path",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_FilePath{
					FilePath: "secrets.txt",
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid file path (absolute)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_FilePath{
					FilePath: "/etc/passwd",
				},
			},
			expectErr: true,
			errMsg:    "invalid secret file path",
		},
		{
			name: "Valid remote content",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: strPtr("https://example.com/secret"),
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Remote content empty URL",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: strPtr(""),
					},
				},
			},
			expectErr: true,
			errMsg:    "remote secret has empty http_url",
		},
		{
			name: "Remote content invalid URL",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: strPtr("not-a-url"),
					},
				},
			},
			expectErr: true,
			errMsg:    "remote secret has invalid http_url",
		},
		{
			name: "Remote content invalid scheme",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: strPtr("ftp://example.com/secret"),
					},
				},
			},
			expectErr: true,
			errMsg:    "remote secret has invalid http_url scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecretValue(tt.secret)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSecretMap(t *testing.T) {
	tests := []struct {
		name      string
		secrets   map[string]*configv1.SecretValue
		expectErr bool
	}{
		{
			name:      "Empty map",
			secrets:   map[string]*configv1.SecretValue{},
			expectErr: false,
		},
		{
			name: "Valid secrets",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": {
					Value: &configv1.SecretValue_FilePath{FilePath: "secret1.txt"},
				},
				"KEY2": {
					Value: &configv1.SecretValue_RemoteContent{
						RemoteContent: &configv1.RemoteContent{HttpUrl: strPtr("https://example.com")},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid secret",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": {
					Value: &configv1.SecretValue_FilePath{FilePath: "/abs/path"},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecretMap(tt.secrets)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateContainerEnvironment_Volumes(t *testing.T) {
	tests := []struct {
		name      string
		env       *configv1.ContainerEnvironment
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid volume",
			env: &configv1.ContainerEnvironment{
				Image: strPtr("alpine"),
				Volumes: map[string]string{
					"./data": "/data",
				},
			},
			expectErr: false,
		},
		{
			name: "Empty host path",
			env: &configv1.ContainerEnvironment{
				Image: strPtr("alpine"),
				Volumes: map[string]string{
					"": "/data",
				},
			},
			expectErr: true,
			errMsg:    "container environment volume host path is empty",
		},
		{
			name: "Empty container path",
			env: &configv1.ContainerEnvironment{
				Image: strPtr("alpine"),
				Volumes: map[string]string{
					"./data": "",
				},
			},
			expectErr: true,
			errMsg:    "container environment volume container path is empty",
		},
		{
			name: "Insecure host path",
			env: &configv1.ContainerEnvironment{
				Image: strPtr("alpine"),
				Volumes: map[string]string{
					"/etc": "/data",
				},
			},
			expectErr: true,
			errMsg:    "not a secure path",
		},
		{
			name: "No image (skip validation)",
			env: &configv1.ContainerEnvironment{
				Image: strPtr(""), // No image
				Volumes: map[string]string{
					"/etc": "/data", // Would be invalid if image was set
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContainerEnvironment(tt.env)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMcpService_StdioConnection(t *testing.T) {
	// Focusing on stdio validation
	tests := []struct {
		name      string
		service   *configv1.McpUpstreamService
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid Stdio",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Command: strPtr("ls"),
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Empty Command",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Command: strPtr(""),
					},
				},
			},
			expectErr: true,
			errMsg:    "has empty command",
		},
		{
			name: "Insecure Working Directory",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Command:          strPtr("ls"),
						WorkingDirectory: strPtr("/etc"),
					},
				},
			},
			expectErr: true,
			errMsg:    "insecure working_directory",
		},
		{
			name: "Invalid Env",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Command: strPtr("ls"),
						Env: map[string]*configv1.SecretValue{
							"BAD": {Value: &configv1.SecretValue_FilePath{FilePath: "/bad"}},
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "invalid secret environment variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMcpService(tt.service)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMcpService_BundleConnection(t *testing.T) {
	tests := []struct {
		name      string
		service   *configv1.McpUpstreamService
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid Bundle",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_BundleConnection{
					BundleConnection: &configv1.McpBundleConnection{
						BundlePath: strPtr("bundle.tar.gz"),
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Empty Bundle Path",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_BundleConnection{
					BundleConnection: &configv1.McpBundleConnection{
						BundlePath: strPtr(""),
					},
				},
			},
			expectErr: true,
			errMsg:    "empty bundle_path",
		},
		{
			name: "Insecure Bundle Path",
			service: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_BundleConnection{
					BundleConnection: &configv1.McpBundleConnection{
						BundlePath: strPtr("/etc/passwd"),
					},
				},
			},
			expectErr: true,
			errMsg:    "insecure bundle_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMcpService(tt.service)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateUpstreamAuthentication(t *testing.T) {
	ctx := context.Background()
	// Mock util.ResolveSecret if needed, or rely on defaults that might fail if not careful.
	// Since util.ResolveSecret is not mocked here and likely reads files or envs, we need to be careful.
	// However, if we pass a plain string to some auth methods (like BearerToken with direct value if supported by proto? No, it uses SecretValue usually? No, the proto has UpstreamBearerTokenAuth with `token` as SecretValue?
	// checking `validateBearerTokenAuth` -> `util.ResolveSecret`.
	// We might need to mock ResolveSecret or set up the environment.
	// Since we can't easily mock ResolveSecret (it's a function in another package), we can test the validation logic structure.

	// But wait, `validateUpstreamAuthentication` calls `util.ResolveSecret`.
	// Let's stick to testing what we can without complex mocking of other packages unless we use a mock framework or interface.
	// Luckily, `validateMtlsAuth` does NOT use `ResolveSecret` for paths, it uses `validation.IsSecurePath`.

	t.Run("MTLS", func(t *testing.T) {
		// Mock osStat
		oldOsStat := osStat
		defer func() { osStat = oldOsStat }()

		// Mock file existence
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // Exists
		}

		// Mock validation.IsSecurePath since it checks real FS for absolute paths if needed or uses logic we can't easily bypass
		// But validation.IsSecurePath is a var in validation package, so we can mock it!
		oldIsSecure := validation.IsSecurePath
		defer func() { validation.IsSecurePath = oldIsSecure }()

		validation.IsSecurePath = func(path string) error {
			if path == "/etc/cert.pem" {
				return assert.AnError
			}
			return nil
		}

		mtls := &configv1.UpstreamAuthentication{
			AuthMethod: &configv1.UpstreamAuthentication_Mtls{
				Mtls: &configv1.UpstreamMTLSAuth{
					ClientCertPath: strPtr("cert.pem"),
					ClientKeyPath:  strPtr("key.pem"),
				},
			},
		}
		err := validateUpstreamAuthentication(ctx, mtls)
		require.NoError(t, err)

		// Test insecure path
		mtls.GetMtls().ClientCertPath = strPtr("/etc/cert.pem")
		err = validateUpstreamAuthentication(ctx, mtls)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a secure path")
	})
}
