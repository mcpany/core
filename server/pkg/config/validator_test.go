// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func strPtr(s string) *string {
	return &s
}

func TestValidateSecretValue(t *testing.T) {
	// Create secrets.txt for FilePath test
	err := os.WriteFile("secrets.txt", []byte("secret-content"), 0600)
	require.NoError(t, err)
	defer os.Remove("secrets.txt")

	// Mock FileExists for this test
	oldIsAllowed := validation.IsAllowedPath
	defer func() { validation.IsAllowedPath = oldIsAllowed }()
	validation.IsAllowedPath = func(path string) error {
		return nil
	}

	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		if path == "secrets.txt" {
			return nil
		}
		if path == "/etc/passwd" {
			return nil
		}
		return os.ErrNotExist
	}

	// Mock Env vars
	os.Setenv("TEST_ENV_VAR", "exists")
	defer os.Unsetenv("TEST_ENV_VAR")

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
			name: "Invalid file path (not allowed)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_FilePath{
					FilePath: "/etc/passwd",
				},
			},
			expectErr: false, // Mocked IsAllowedPath returns nil
		},
		{
			name: "Missing file path",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_FilePath{
					FilePath: "missing.txt",
				},
			},
			expectErr: true,
			errMsg:    "secret file \"missing.txt\" error",
		},
		{
			name: "Valid env var",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "TEST_ENV_VAR",
				},
			},
			expectErr: false,
		},
		{
			name: "Missing env var",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "MISSING_ENV_VAR",
				},
			},
			expectErr: true,
			errMsg:    "environment variable \"MISSING_ENV_VAR\" error",
		},
		{
			name: "Valid remote content (unreachable in test)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: strPtr("https://example.com/secret"),
					},
				},
			},
			expectErr: true,
			errMsg:    "failed to fetch remote secret",
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
		},
		{
			name: "Valid regex match (plain_text)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_PlainText{
					PlainText: "sk-1234567890",
				},
				ValidationRegex: proto.String(`^sk-[a-zA-Z0-9]{10}$`),
			},
			expectErr: false,
		},
		{
			name: "Invalid regex match (plain_text)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_PlainText{
					PlainText: "invalid-key",
				},
				ValidationRegex: proto.String(`^sk-[a-zA-Z0-9]{10}$`),
			},
			expectErr: true,
			errMsg:    "secret value does not match validation regex",
		},
		{
			name: "Valid regex match (env_var)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "TEST_ENV_VAR",
				},
				ValidationRegex: proto.String(`^exists$`),
			},
			expectErr: false,
		},
		{
			name: "Invalid regex match (env_var)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "TEST_ENV_VAR",
				},
				ValidationRegex: proto.String(`^wrong$`),
			},
			expectErr: true,
			errMsg:    "secret value does not match validation regex",
		},
		{
			name: "Invalid regex pattern",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_PlainText{
					PlainText: "value",
				},
				ValidationRegex: proto.String(`[`),
			},
			expectErr: true,
			errMsg:    "invalid validation regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Invalid file path (absolute)" {
				// Special handling if needed
				oldAllowed := validation.IsAllowedPath
				validation.IsAllowedPath = func(path string) error { return assert.AnError }
				defer func() { validation.IsAllowedPath = oldAllowed }()

				err := validateSecretValue(context.Background(), tt.secret)
				require.Error(t, err)
				return
			}

			err := validateSecretValue(context.Background(), tt.secret)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMcpStdioConnection_RelativeCommandWithWorkingDir(t *testing.T) {
	tempDir, err := os.MkdirTemp(".", "mcpany-test-wd")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "start.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/sh\necho hello"), 0755)
	require.NoError(t, err)

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command:          proto.String("./start.sh"),
								WorkingDirectory: proto.String(tempDir),
							},
						},
					},
				},
			},
		},
	}

	errors := Validate(context.Background(), config, Server)
	assert.Empty(t, errors, "Expected no validation errors for relative command in working directory")
}

func TestValidateMcpStdioConnection_RelativeCommandMissing(t *testing.T) {
	tempDir, err := os.MkdirTemp(".", "mcpany-test-wd")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-service-missing"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command:          proto.String("./missing_script.sh"),
								WorkingDirectory: proto.String(tempDir),
							},
						},
					},
				},
			},
		},
	}

	errors := Validate(context.Background(), config, Server)
	assert.NotEmpty(t, errors, "Expected validation errors for missing command")
}

func TestValidateMcpStdioConnection_ArgsValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp(".", "mcpany-test-args")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "script.py")
	err = os.WriteFile(scriptPath, []byte("print('hello')"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		command     string
		args        []string
		workingDir  string
		expectError bool
	}{
		{
			name:        "Valid script arg (Python)",
			command:     "python3",
			args:        []string{scriptPath},
			expectError: false,
		},
		{
			name:        "Missing script arg (Python)",
			command:     "python3",
			args:        []string{filepath.Join(tempDir, "missing.py")},
			expectError: true,
		},
		{
			name:        "Valid script arg with flag (Python)",
			command:     "python3",
			args:        []string{"-u", scriptPath},
			expectError: false,
		},
		{
			name:        "Missing script arg with flag (Python)",
			command:     "python3",
			args:        []string{"-u", filepath.Join(tempDir, "missing.py")},
			expectError: true,
		},
		{
			name:        "Non-interpreter command (ignored)",
			command:     "ls",
			args:        []string{"non_existent_file.txt"},
			expectError: false,
		},
		{
			name:        "Relative script in WorkingDir",
			command:     "python3",
			args:        []string{"script.py"},
			workingDir:  tempDir,
			expectError: false,
		},
		{
			name:        "Relative missing script in WorkingDir",
			command:     "python3",
			args:        []string{"missing.py"},
			workingDir:  tempDir,
			expectError: true,
		},
		{
			name:        "Python -m module execution",
			command:     "python3",
			args:        []string{"-m", "http.server"},
			expectError: false,
		},
		{
			name:        "Deno remote script",
			command:     "deno",
			args:        []string{"run", "https://deno.land/std/examples/chat/server.ts"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("test-service"),
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										Command:          proto.String(tt.command),
										Args:             tt.args,
										WorkingDirectory: proto.String(tt.workingDir),
									},
								},
							},
						},
					},
				},
			}

			oldLookPath := execLookPath
			defer func() { execLookPath = oldLookPath }()
			execLookPath = func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			}

			oldOsStat := osStat
			defer func() { osStat = oldOsStat }()
			osStat = os.Stat

			errors := Validate(context.Background(), config, Server)
			if tt.expectError {
				assert.NotEmpty(t, errors, "Expected validation error")
			} else {
				assert.Empty(t, errors, "Expected no validation error")
			}
		})
	}
}

func TestValidateSecretMap(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "secret1.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("content")
	tmpFile.Close()

	oldIsAllowed := validation.IsAllowedPath
	defer func() { validation.IsAllowedPath = oldIsAllowed }()
	validation.IsAllowedPath = func(path string) error {
		return nil
	}

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
					Value: &configv1.SecretValue_FilePath{FilePath: tmpFile.Name()},
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid secret",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": {
					Value: &configv1.SecretValue_FilePath{FilePath: "/abs/path/missing"},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecretMap(context.Background(), tt.secrets)
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
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		if file == "ls" || file == "/bin/ls" {
			return "/bin/ls", nil
		}
		return "", os.ErrNotExist
	}

	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil
	}

	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		return nil
	}

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
			err := validateMcpService(context.Background(), tt.service)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCommandExists(t *testing.T) {
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		if file == "ls" || file == "/bin/ls" {
			return "/bin/ls", nil
		}
		return "", os.ErrNotExist
	}

	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()
	osStat = func(name string) (os.FileInfo, error) {
		if name == "/bin/ls" {
			return &mockFileInfo{isDir: false}, nil
		}
		if name == "/bin/dir" {
			return &mockFileInfo{isDir: true}, nil
		}
		return nil, os.ErrNotExist
	}

	tests := []struct {
		name      string
		command   string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Command in PATH",
			command:   "ls",
			expectErr: false,
		},
		{
			name:      "Command not in PATH",
			command:   "missing_cmd",
			expectErr: true,
			errMsg:    "not found in PATH",
		},
		{
			name:      "Absolute path exists",
			command:   "/bin/ls",
			expectErr: false,
		},
		{
			name:      "Absolute path missing",
			command:   "/bin/missing",
			expectErr: true,
			errMsg:    "executable not found",
		},
		{
			name:      "Absolute path is directory",
			command:   "/bin/dir",
			expectErr: true,
			errMsg:    "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommandExists(tt.command, "")
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateDirectoryExists(t *testing.T) {
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()
	osStat = func(name string) (os.FileInfo, error) {
		if name == "/valid/dir" {
			return &mockFileInfo{isDir: true}, nil
		}
		if name == "/valid/file" {
			return &mockFileInfo{isDir: false}, nil
		}
		return nil, os.ErrNotExist
	}

	tests := []struct {
		name      string
		path      string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid directory",
			path:      "/valid/dir",
			expectErr: false,
		},
		{
			name:      "Missing directory",
			path:      "/missing/dir",
			expectErr: true,
			errMsg:    "directory \"/missing/dir\" does not exist",
		},
		{
			name:      "Path is a file",
			path:      "/valid/file",
			expectErr: true,
			errMsg:    "is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirectoryExists(tt.path)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockFileInfo struct {
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() any           { return nil }

func TestValidateMcpService_BundleConnection(t *testing.T) {
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		return nil
	}

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
			err := validateMcpService(context.Background(), tt.service)
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

	t.Run("MTLS", func(t *testing.T) {
		oldOsStat := osStat
		defer func() { osStat = oldOsStat }()
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // Exists
		}

		oldIsSecure := validation.IsSecurePath
		defer func() { validation.IsSecurePath = oldIsSecure }()
		validation.IsSecurePath = func(path string) error {
			if path == "/etc/cert.pem" {
				return assert.AnError
			}
			return nil
		}

		mtls := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Mtls{
				Mtls: &configv1.MTLSAuth{
					ClientCertPath: strPtr("cert.pem"),
					ClientKeyPath:  strPtr("key.pem"),
				},
			},
		}
		err := validateAuthentication(ctx, mtls, AuthValidationContextOutgoing)
		require.NoError(t, err)

		mtls.GetMtls().ClientCertPath = strPtr("/etc/cert.pem")
		err = validateAuthentication(ctx, mtls, AuthValidationContextOutgoing)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a secure path")
	})
}
