// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidateSecretValue_Coverage(t *testing.T) {
	// Mock FileExists to pass existing tests
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		return nil
	}

	tests := []struct {
		name    string
		secret  *configv1.SecretValue
		wantErr bool
	}{
		{
			name:    "nil secret",
			secret:  nil,
			wantErr: false,
		},
		{
			name: "valid file path",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_FilePath{
					FilePath: "secrets/key.txt",
				},
			},
			wantErr: false, // assuming "secrets/key.txt" is considered relative and secure
		},
		{
			name: "invalid file path (absolute)",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_FilePath{
					FilePath: "/etc/passwd",
				},
			},
			wantErr: true,
		},
		{
			name: "valid remote content",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: ptr("https://example.com/secret"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "remote content empty url",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: ptr(""),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "remote content invalid url",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: ptr("://invalid"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "remote content invalid scheme",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_RemoteContent{
					RemoteContent: &configv1.RemoteContent{
						HttpUrl: ptr("ftp://example.com/secret"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid environment variable",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "PATH",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid environment variable",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "NON_EXISTENT_ENV_VAR_12345",
				},
			},
			wantErr: true,
		},
	}

	// Ensure PATH exists
	os.Setenv("PATH", "/bin")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecretValue(tt.secret)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSecretMap_Coverage(t *testing.T) {
	// Mock FileExists to pass existing tests
	oldFileExists := validation.FileExists
	defer func() { validation.FileExists = oldFileExists }()
	validation.FileExists = func(path string) error {
		return nil
	}

	tests := []struct {
		name    string
		secrets map[string]*configv1.SecretValue
		wantErr bool
	}{
		{
			name:    "empty map",
			secrets: map[string]*configv1.SecretValue{},
			wantErr: false,
		},
		{
			name: "valid map",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": {
					Value: &configv1.SecretValue_FilePath{
						FilePath: "secret1.txt",
					},
				},
				"KEY2": {
					Value: &configv1.SecretValue_RemoteContent{
						RemoteContent: &configv1.RemoteContent{
							HttpUrl: ptr("https://example.com"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid map entry",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": {
					Value: &configv1.SecretValue_FilePath{
						FilePath: "/absolute/path",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecretMap(tt.secrets)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
