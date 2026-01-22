// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecret_ValidationRegex(t *testing.T) {
	ctx := context.Background()

	t.Setenv("TEST_REGEX_ENV", "env-value-123")

	tests := []struct {
		name        string
		secret      *configv1.SecretValue
		expectError bool
	}{
		{
			name: "PlainText valid regex",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_PlainText{
					PlainText: "valid-key-123",
				},
				ValidationRegex: proto.String("^valid-key-[0-9]{3}$"),
			},
			expectError: false,
		},
		{
			name: "PlainText invalid regex",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_PlainText{
					PlainText: "invalid-key",
				},
				ValidationRegex: proto.String("^valid-key-[0-9]{3}$"),
			},
			expectError: true, // Should fail if validation is applied
		},
		{
			name: "EnvironmentVariable valid regex",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "TEST_REGEX_ENV",
				},
				ValidationRegex: proto.String("^env-value-[0-9]{3}$"),
			},
			expectError: false,
		},
		{
			name: "EnvironmentVariable invalid regex",
			secret: &configv1.SecretValue{
				Value: &configv1.SecretValue_EnvironmentVariable{
					EnvironmentVariable: "TEST_REGEX_ENV",
				},
				ValidationRegex: proto.String("^wrong-pattern$"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ResolveSecret(ctx, tt.secret)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "does not match validation regex")
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, val)
			}
		})
	}
}
