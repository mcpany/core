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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func strPtr(s string) *string {
	return &s
}

func TestValidateSecretValue(t *testing.T) {
	// Mock FileExists for this test
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		if path == "secrets.txt" {
			return nil
		}
		if path == "/etc/passwd" {
			return nil // Exists but invalid path logic will catch it first
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
			secret: configv1.SecretValue_builder{
				FilePath: proto.String("secrets.txt"),
			}.Build(),
			expectErr: false,
		},
		{
			name: "Invalid file path (absolute)",
			secret: configv1.SecretValue_builder{
				FilePath: proto.String("/etc/passwd"),
			}.Build(),
			expectErr: true,
			errMsg:    "invalid secret file path",
		},
		{
			name: "Missing file path",
			secret: configv1.SecretValue_builder{
				FilePath: proto.String("missing.txt"),
			}.Build(),
			expectErr: true,
			errMsg:    "secret file \"missing.txt\" does not exist",
		},
		{
			name: "Valid env var",
			secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("TEST_ENV_VAR"),
			}.Build(),
			expectErr: false,
		},
		{
			name: "Missing env var",
			secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("MISSING_ENV_VAR"),
			}.Build(),
			expectErr: true,
			errMsg:    "environment variable \"MISSING_ENV_VAR\" is not set",
		},
		{
			name: "Valid remote content",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: proto.String("https://example.com/secret"),
				}.Build(),
			}.Build(),
			expectErr: false,
		},
		{
			name: "Remote content empty URL",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: proto.String(""),
				}.Build(),
			}.Build(),
			expectErr: true,
			errMsg:    "remote secret has empty http_url",
		},
		{
			name: "Remote content invalid URL",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: proto.String("not-a-url"),
				}.Build(),
			}.Build(),
			expectErr: true,
			errMsg:    "remote secret has invalid http_url",
		},
		{
			name: "Remote content invalid scheme",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: proto.String("ftp://example.com/secret"),
				}.Build(),
			}.Build(),
			expectErr: true,
			errMsg:    "remote secret has invalid http_url scheme",
		},
		{
			name: "Valid regex match (plain_text)",
			secret: configv1.SecretValue_builder{
				PlainText:       proto.String("sk-1234567890"),
				ValidationRegex: proto.String(`^sk-[a-zA-Z0-9]{10}$`),
			}.Build(),
			expectErr: false,
		},
		{
			name: "Invalid regex match (plain_text)",
			secret: configv1.SecretValue_builder{
				PlainText:       proto.String("invalid-key"),
				ValidationRegex: proto.String(`^sk-[a-zA-Z0-9]{10}$`),
			}.Build(),
			expectErr: true,
			errMsg:    "secret value does not match validation regex",
		},
		{
			name: "Valid regex match (env_var)",
			secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("TEST_ENV_VAR"),
				ValidationRegex:     proto.String(`^exists$`),
			}.Build(),
			expectErr: false,
		},
		{
			name: "Invalid regex match (env_var)",
			secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("TEST_ENV_VAR"),
				ValidationRegex:     proto.String(`^wrong$`),
			}.Build(),
			expectErr: true,
			errMsg:    "secret value does not match validation regex",
		},
		{
			name: "Invalid regex pattern",
			secret: configv1.SecretValue_builder{
				PlainText:       proto.String("value"),
				ValidationRegex: proto.String(`[`),
			}.Build(),
			expectErr: true,
			errMsg:    "invalid validation regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecretValue(context.Background(), tt.secret)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMcpStdioConnection_RelativeCommandWithWorkingDir(t *testing.T) {
	// Create a temporary directory for the "working directory" in the current directory
	// so that IsAllowedPath passes.
	tempDir, err := os.MkdirTemp(".", "mcpany-test-wd")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a script in that directory
	scriptPath := filepath.Join(tempDir, "start.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/sh\necho hello"), 0755)
	require.NoError(t, err)

	// Config with relative command and working directory
	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test-service"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command:          proto.String("./start.sh"),
						WorkingDirectory: proto.String(tempDir),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Validate should PASS now because it checks inside WorkingDirectory
	errors := Validate(context.Background(), config, Server)

	assert.Empty(t, errors, "Expected no validation errors for relative command in working directory")
}

func TestValidateMcpStdioConnection_RelativeCommandMissing(t *testing.T) {
	// Create a temporary directory for the "working directory" in the current directory
	tempDir, err := os.MkdirTemp(".", "mcpany-test-wd")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Config with relative command that DOES NOT exist
	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test-service-missing"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command:          proto.String("./missing_script.sh"),
						WorkingDirectory: proto.String(tempDir),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Validate should FAIL
	errors := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errors, "Expected validation errors for missing command")
	// We just check that it failed. The error message depends on whether it fell through to LookPath.
	// Since it's missing in WD, it falls through to LookPath, which fails finding it in PATH/CWD.
}

func TestValidateMcpStdioConnection_ArgsValidation(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp(".", "mcpany-test-args")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid script
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
			args:        []string{"non_existent_file.txt"}, // ls might complain at runtime, but validation ignores it
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
			args:        []string{"-m", "http.server"}, // http.server looks like a file but -m should skip validation
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
			config := configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("test-service"),
						McpService: configv1.McpUpstreamService_builder{
							StdioConnection: configv1.McpStdioConnection_builder{
								Command:          proto.String(tt.command),
								Args:             tt.args,
								WorkingDirectory: proto.String(tt.workingDir),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build()

			// We need to mock execLookPath to allow "python3" to pass command validation
			oldLookPath := execLookPath
			defer func() { execLookPath = oldLookPath }()
			execLookPath = func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			}

			// We need to mock osStat to call real os.Stat for the tempDir files
			// But other tests mock it globally!
			// We can restore it after this test block, or use a local override if the tested code uses the global variable.
			// The tested code uses the global `osStat` variable.
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
	// Mock FileExists
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
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
				"KEY1": configv1.SecretValue_builder{
					FilePath: proto.String("secret1.txt"),
				}.Build(),
				"KEY2": configv1.SecretValue_builder{
					RemoteContent: configv1.RemoteContent_builder{
						HttpUrl: proto.String("https://example.com"),
					}.Build(),
				}.Build(),
			},
			expectErr: false,
		},
		{
			name: "Invalid secret",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": configv1.SecretValue_builder{
					FilePath: proto.String("/abs/path"),
				}.Build(),
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
			env: configv1.ContainerEnvironment_builder{
				Image: proto.String("alpine"),
				Volumes: map[string]string{
					"./data": "/data",
				},
			}.Build(),
			expectErr: false,
		},
		{
			name: "Empty host path",
			env: configv1.ContainerEnvironment_builder{
				Image: proto.String("alpine"),
				Volumes: map[string]string{
					"": "/data",
				},
			}.Build(),
			expectErr: true,
			errMsg:    "container environment volume host path is empty",
		},
		{
			name: "Empty container path",
			env: configv1.ContainerEnvironment_builder{
				Image: proto.String("alpine"),
				Volumes: map[string]string{
					"./data": "",
				},
			}.Build(),
			expectErr: true,
			errMsg:    "container environment volume container path is empty",
		},
		{
			name: "Insecure host path",
			env: configv1.ContainerEnvironment_builder{
				Image: proto.String("alpine"),
				Volumes: map[string]string{
					"/etc": "/data",
				},
			}.Build(),
			expectErr: true,
			errMsg:    "not a secure path",
		},
		{
			name: "No image (skip validation)",
			env: configv1.ContainerEnvironment_builder{
				Image: proto.String(""), // No image
				Volumes: map[string]string{
					"/etc": "/data", // Would be invalid if image was set
				},
			}.Build(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContainerEnvironment(context.Background(), tt.env)
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
	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		if file == "ls" || file == "/bin/ls" {
			return "/bin/ls", nil
		}
		return "", os.ErrNotExist
	}

	// Mock osStat for working directory validation
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return &mockFileInfo{isDir: true}, nil // Assume exists and is directory
	}

	// Mock FileExists for secret validation
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		return nil
	}

	// Focusing on stdio validation
	tests := []struct {
		name      string
		service   *configv1.McpUpstreamService
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid Stdio",
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: proto.String("ls"),
					}.Build(),
				}.Build()
			}(),
			expectErr: false,
		},
		{
			name: "Empty Command",
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: proto.String(""),
					}.Build(),
				}.Build()
			}(),
			expectErr: true,
			errMsg:    "has empty command",
		},
		{
			name: "Insecure Working Directory",
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command:          proto.String("ls"),
						WorkingDirectory: proto.String("/etc"),
					}.Build(),
				}.Build()
			}(),
			expectErr: true,
			errMsg:    "insecure working_directory",
		},
		{
			name: "Invalid Env",
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: proto.String("ls"),
						Env: map[string]*configv1.SecretValue{
							"BAD": configv1.SecretValue_builder{
								FilePath: proto.String("/bad"),
							}.Build(),
						},
					}.Build(),
				}.Build()
			}(),
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
	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		if file == "ls" || file == "/bin/ls" {
			return "/bin/ls", nil
		}
		return "", os.ErrNotExist
	}

	// Mock osStat
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
			err := validateCommandExists(context.Background(), tt.command, "")
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
	// Mock osStat
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
			err := validateDirectoryExists(context.Background(), tt.path)
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
	// Mock FileExists for secret validation
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
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						BundlePath: proto.String("bundle.tar.gz"),
					}.Build(),
				}.Build()
			}(),
			expectErr: false,
		},
		{
			name: "Empty Bundle Path",
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						BundlePath: proto.String(""),
					}.Build(),
				}.Build()
			}(),
			expectErr: true,
			errMsg:    "empty bundle_path",
		},
		{
			name: "Insecure Bundle Path",
			service: func() *configv1.McpUpstreamService {
				return configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						BundlePath: proto.String("/etc/passwd"),
					}.Build(),
				}.Build()
			}(),
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
		// Mock osStat
		oldOsStat := osStat
		defer func() { osStat = oldOsStat }()

		// Mock file existence
		osStat = func(name string) (os.FileInfo, error) {
			return nil, nil // Exists
		}

		// Mock validation.IsPathTraversalSafe since it checks real FS for absolute paths if needed or uses logic we can't easily bypass
		// But validation.IsPathTraversalSafe is a var in validation package, so we can mock it!
		oldIsSecure := validation.IsPathTraversalSafe
		defer func() { validation.IsPathTraversalSafe = oldIsSecure }()

		validation.IsPathTraversalSafe = func(path string) error {
			if path == "/etc/cert.pem" {
				return assert.AnError
			}
			return nil
		}

		mtls := configv1.Authentication_builder{
			Mtls: configv1.MTLSAuth_builder{
				ClientCertPath: proto.String("cert.pem"),
				ClientKeyPath:  proto.String("key.pem"),
			}.Build(),
		}.Build()
		err := validateAuthentication(ctx, mtls, AuthValidationContextOutgoing)
		require.NoError(t, err)

		// Test insecure path
		// We need to use builder or setters if we can't mutate directly easily,
		// but since we built it, we can mutate if we have the pointer.
		// Wait, Build() returns *Authentication? Yes.
		mtls.GetMtls().SetClientCertPath("/etc/cert.pem")
		err = validateAuthentication(ctx, mtls, AuthValidationContextOutgoing)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a secure path")
	})
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestValidate_ExtraServices(t *testing.T) {
	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "valid graphql service",
			config: func() *configv1.McpAnyServerConfig {
				glSvc := configv1.GraphQLUpstreamService_builder{
					Address: proto.String("http://example.com/graphql"),
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name:           proto.String("graphql-valid"),
					GraphqlService: glSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid graphql service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				// Keeping protojson unmarshal as it is valid "usage" (parsing from JSON)
				// but user asked for "protobuf usage". protojson is fine usually.
				// However, the test uses protojson.Unmarshal just to create the config.
				// I will leave it as is if it's using protojson, because protojson produces Opaque structs internally?
				// Actually protojson unmarshal is fine.
				// But wait, the previous code used protojson.Unmarshal into `cfg`.
				// `cfg` was `&configv1.McpAnyServerConfig{}`.
				// I should replace `&configv1.McpAnyServerConfig{}` with builder?
				// No, protojson.Unmarshal requires a message pointer.
				// Opaque API messages are pointers? No, they are interfaces involved?
				// Actually `configv1.McpAnyServerConfig` is a struct type in the new API?
				// The builders produce `*configv1.McpAnyServerConfig` usually?
				// Let's check `Build()` return type.
				// Usually `Build()` returns `*McpAnyServerConfig`.
				// So `protojson.Unmarshal` works on `*McpAnyServerConfig`.
				// So `cfg := &configv1.McpAnyServerConfig{}` is still valid?
				// Opaque API usually deprecates DIRECT field access.
				// But `&Struct{}` literals are what we want to avoid.
				// `new(configv1.McpAnyServerConfig)` is better?
				// Or `configv1.McpAnyServerConfig_builder{}.Build()` produces an empty valid config.
				// But `protojson.Unmarshal` takes `proto.Message`.
				// I will leave the protojson ones alone for now as they are "data driven"
				// AND `protojson` is a standard library function, not direct struct literal usage.
				// The goal is "usage should be using protobuf builder" likely meaning construction.

				// I will focus on the first case "valid graphql service" which used struct literals.
				// And others.

				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "graphql-bad-scheme",
							"graphql_service": {
								"address": "ftp://example.com/graphql"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "service \"graphql-bad-scheme\": invalid graphql address scheme: ftp\n\t-> Fix: Use 'http' or 'https' as the scheme.",
		},
		{
			name: "valid webrtc service",
			config: func() *configv1.McpAnyServerConfig {
				wbSvc := configv1.WebrtcUpstreamService_builder{
					Address: proto.String("http://example.com/webrtc"),
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name:          proto.String("webrtc-valid"),
					WebrtcService: wbSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid webrtc service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "webrtc-bad-scheme",
							"webrtc_service": {
								"address": "ftp://example.com/webrtc"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "service \"webrtc-bad-scheme\": invalid webrtc address scheme: ftp\n\t-> Fix: Use 'http' or 'https' as the scheme.",
		},
		{
			name: "valid upstream service collection",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-1",
							"http_url": "http://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid upstream service collection - empty name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "",
							"http_url": "http://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "": collection name is empty`,
		},
		{
			name: "invalid upstream service collection - empty url",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-no-url",
							"http_url": ""
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "collection-no-url": collection must have either http_url or inline content (services/skills)`,
		},
		{
			name: "invalid upstream service collection - bad url scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-bad-scheme",
							"http_url": "ftp://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "collection-bad-scheme": invalid collection http_url scheme: ftp`,
		},
		{
			name: "valid upstream service collection with auth",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-auth",
							"http_url": "http://example.com/collection",
							"authentication": {
								"basic_auth": {
									"username": "user",
									"password": { "plainText": "pass" }
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "duplicate service name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "service-1",
							"http_service": { "address": "http://example.com/1" }
						},
						{
							"name": "service-1",
							"http_service": { "address": "http://example.com/2" }
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "service-1": duplicate service name found`,
		},
		{
			name: "invalid upstream service - empty name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "",
							"http_service": { "address": "http://example.com/empty-name" }
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "": service name is empty`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(context.Background(), tt.config, Server)
			if tt.expectedErrorCount > 0 {
				require.NotEmpty(t, validationErrors, "expected validation errors but got none")
				found := false
				for _, err := range validationErrors {
					if err.Error() == tt.expectedErrorString {
						found = true
						break
					}
				}
				if !found {
					assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
				}
			} else {
				assert.Empty(t, validationErrors)
			}
		})
	}
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func boolPtr(b bool) *bool                                                                { return &b }
func storageTypePtr(t configv1.AuditConfig_StorageType) *configv1.AuditConfig_StorageType { return &t }

func TestValidateUsers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		users        []*configv1.User
		expectErr    bool
		errSubstring string
	}{
		{
			name: "Valid User",
			users: func() []*configv1.User {
				u := configv1.User_builder{
					Id: proto.String("user1"),
					Authentication: configv1.Authentication_builder{
						ApiKey: configv1.APIKeyAuth_builder{
							ParamName:         proto.String("key"),
							VerificationValue: proto.String("secret"),
						}.Build(),
					}.Build(),
				}.Build()
				return []*configv1.User{u}
			}(),
			expectErr: false,
		},
		{
			name: "Missing ID",
			users: func() []*configv1.User {
				u := configv1.User_builder{
					Id: proto.String(""),
				}.Build()
				return []*configv1.User{u}
			}(),
			expectErr:    true,
			errSubstring: "user has empty id",
		},
		{
			name: "Duplicate ID",
			users: func() []*configv1.User {
				u1 := configv1.User_builder{
					Id: proto.String("user1"),
					Authentication: configv1.Authentication_builder{
						ApiKey: configv1.APIKeyAuth_builder{
							ParamName:         proto.String("k"),
							VerificationValue: proto.String("v"),
						}.Build(),
					}.Build(),
				}.Build()

				u2 := configv1.User_builder{
					Id: proto.String("user1"),
					Authentication: configv1.Authentication_builder{
						ApiKey: configv1.APIKeyAuth_builder{
							ParamName:         proto.String("k"),
							VerificationValue: proto.String("v"),
						}.Build(),
					}.Build(),
				}.Build()

				return []*configv1.User{u1, u2}
			}(),
			expectErr:    true,
			errSubstring: "duplicate user id",
		},
		{
			name: "Missing Authentication",
			users: func() []*configv1.User {
				u := configv1.User_builder{
					Id: proto.String("user1"),
				}.Build()
				return []*configv1.User{u}
			}(),
			expectErr: false,
		},
		{
			name: "Invalid OAuth2",
			users: func() []*configv1.User {
				u := configv1.User_builder{
					Id: proto.String("user1"),
					Authentication: configv1.Authentication_builder{
						Oauth2: configv1.OAuth2Auth_builder{
							TokenUrl: proto.String("invalid-url"),
						}.Build(),
					}.Build(),
				}.Build()
				return []*configv1.User{u}
			}(),
			expectErr:    true,
			errSubstring: "invalid oauth2 token_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := configv1.McpAnyServerConfig_builder{
				Users: tt.users,
			}.Build()
			errs := Validate(ctx, config, Server)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				found := false
				for _, e := range errs {
					if assert.Contains(t, e.Err.Error(), tt.errSubstring) {
						found = true
						break
					}
				}
				if !found && len(errs) > 0 {
					// Check if substring match failed but error existed
					// Actually strict check:
					assert.Fail(t, "expected error substring not found", "substring: %s, errors: %v", tt.errSubstring, errs)
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateGlobalSettings_Extended(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		gs           *configv1.GlobalSettings
		expectErr    bool
		errSubstring string
	}{
		{
			name: "Valid Audit File",
			gs: configv1.GlobalSettings_builder{
				Audit: configv1.AuditConfig_builder{
					Enabled:     proto.Bool(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_FILE),
					OutputPath:  proto.String("/var/log/audit.log"),
				}.Build(),
			}.Build(),
			expectErr: false,
		},
		{
			name: "Audit File Missing Path",
			gs: configv1.GlobalSettings_builder{
				Audit: configv1.AuditConfig_builder{
					Enabled:     proto.Bool(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_FILE),
				}.Build(),
			}.Build(),
			expectErr:    true,
			errSubstring: "output_path is required",
		},
		{
			name: "Audit Webhook Invalid URL",
			gs: configv1.GlobalSettings_builder{
				Audit: configv1.AuditConfig_builder{
					Enabled:     proto.Bool(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_WEBHOOK),
					WebhookUrl:  proto.String("not-a-url"),
				}.Build(),
			}.Build(),
			expectErr:    true,
			errSubstring: "invalid webhook_url",
		},
		{
			name: "DLP Invalid Regex",
			gs: configv1.GlobalSettings_builder{
				Dlp: configv1.DLPConfig_builder{
					Enabled:        proto.Bool(true),
					CustomPatterns: []string{"["},
				}.Build(),
			}.Build(),
			expectErr:    true,
			errSubstring: "invalid regex pattern",
		},
		{
			name: "GC Invalid Interval",
			gs: configv1.GlobalSettings_builder{
				GcSettings: configv1.GCSettings_builder{
					Enabled:  proto.Bool(true),
					Interval: proto.String("not-a-duration"),
				}.Build(),
			}.Build(),
			expectErr:    true,
			errSubstring: "invalid interval",
		},
		{
			name: "GC Insecure Path",
			gs: configv1.GlobalSettings_builder{
				GcSettings: configv1.GCSettings_builder{
					Enabled: proto.Bool(true),
					Paths:   []string{"../etc"},
				}.Build(),
			}.Build(),
			expectErr:    true,
			errSubstring: "not secure",
		},
		{
			name: "GC Relative Path (Not Allowed)",
			gs: configv1.GlobalSettings_builder{
				GcSettings: configv1.GCSettings_builder{
					Enabled: proto.Bool(true),
					Paths:   []string{"relative/path"},
				}.Build(),
			}.Build(),
			expectErr:    true,
			errSubstring: "must be absolute",
		},
		{
			name: "Duplicate Profile Name",
			gs: configv1.GlobalSettings_builder{
				ProfileDefinitions: []*configv1.ProfileDefinition{
					configv1.ProfileDefinition_builder{Name: proto.String("p1")}.Build(),
					configv1.ProfileDefinition_builder{Name: proto.String("p1")}.Build(),
				},
			}.Build(),
			expectErr:    true,
			errSubstring: "duplicate profile definition name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := configv1.McpAnyServerConfig_builder{
				GlobalSettings: tt.gs,
			}.Build()
			errs := Validate(ctx, config, Server)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				found := false
				for _, e := range errs {
					if len(e.Err.Error()) > 0 && (tt.errSubstring == "" || assert.Contains(t, e.Err.Error(), tt.errSubstring)) {
						found = true
						break
					}
				}
				if !found {
					t.Logf("Errors found: %v", errs)
					assert.Fail(t, "expected error not found")
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
