// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestValidateSecretValue_Coverage(t *testing.T) {
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
			secret: configv1.SecretValue_builder{
				FilePath: ptr("secrets/key.txt"),
			}.Build(),
			wantErr: false, // assuming "secrets/key.txt" is considered relative and secure
		},
		{
			name: "invalid file path (absolute)",
			secret: configv1.SecretValue_builder{
				FilePath: ptr("/etc/passwd"),
			}.Build(),
			wantErr: true,
		},
		{
			name: "valid remote content",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: ptr("https://example.com/secret"),
				}.Build(),
			}.Build(),
			wantErr: false,
		},
		{
			name: "remote content empty url",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: ptr(""),
				}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name: "remote content invalid url",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: ptr("://invalid"),
				}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name: "remote content invalid scheme",
			secret: configv1.SecretValue_builder{
				RemoteContent: configv1.RemoteContent_builder{
					HttpUrl: ptr("ftp://example.com/secret"),
				}.Build(),
			}.Build(),
			wantErr: true,
		},
	}

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
				"KEY1": configv1.SecretValue_builder{
					FilePath: ptr("secret1.txt"),
				}.Build(),
				"KEY2": configv1.SecretValue_builder{
					RemoteContent: configv1.RemoteContent_builder{
						HttpUrl: ptr("https://example.com"),
					}.Build(),
				}.Build(),
			},
			wantErr: false,
		},
		{
			name: "invalid map entry",
			secrets: map[string]*configv1.SecretValue{
				"KEY1": configv1.SecretValue_builder{
					FilePath: ptr("/absolute/path"),
				}.Build(),
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
