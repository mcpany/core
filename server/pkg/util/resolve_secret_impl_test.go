// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecretImplCoverage(t *testing.T) {
	ctx := context.Background()

	t.Run("RecursionDepthExceeded", func(t *testing.T) {
		var recursiveSecret *configv1.SecretValue
		for i := 0; i < 15; i++ {
			if i == 0 {
				recursiveSecret = configv1.SecretValue_builder{
					RemoteContent: configv1.RemoteContent_builder{
						HttpUrl: proto.String("http://localhost"),
					}.Build(),
				}.Build()
			} else {
				recursiveSecret = configv1.SecretValue_builder{
					RemoteContent: configv1.RemoteContent_builder{
						HttpUrl: proto.String("http://localhost"),
						Auth: configv1.AuthMethod_builder{
							ApiKey: configv1.ApiKeyAuth_builder{
								ParamName: proto.String("key"),
								Value:     recursiveSecret,
							}.Build(),
						}.Build(),
					}.Build(),
				}.Build()
			}
		}

		_, err := util.ResolveSecret(ctx, recursiveSecret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret resolution exceeded max recursion depth")
	})

	t.Run("EnvVar_NotAllowed", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NOT_ALLOWED_VAR_XYZ"),
		}.Build()

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access to environment variable")
	})

	t.Run("FilePath_InvalidPath", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			FilePath: proto.String("/etc/shadow"),
		}.Build()

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid secret file path")
	})

	t.Run("FilePath_ReadError", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "not-exists.txt")

		util.SetAllowedPathsForTest([]string{t.TempDir()})
		t.Cleanup(func() { util.SetAllowedPathsForTest(nil) })

		secret := configv1.SecretValue_builder{
			FilePath: proto.String(tmpFile),
		}.Build()

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read secret from file")
	})

	t.Run("AwsSecretManager_BadConfig", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			AwsSecretManager: configv1.AwsSecretManager_builder{
				SecretId: proto.String("my-secret"),
				Region: proto.String("us-east-1"),
				Profile: proto.String("invalid-profile"),
			}.Build(),
		}.Build()

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
	})
}

func TestResolveSecretImplCoverageExtended(t *testing.T) {
	ctx := context.Background()

	t.Run("PlainText_RegexValidationMismatch", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			PlainText: proto.String("my-secret"),
			ValidationRegex: proto.String("^[0-9]+$"),
		}.Build()

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret value does not match validation regex")
	})

	t.Run("PlainText_InvalidRegex", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			PlainText: proto.String("my-secret"),
			ValidationRegex: proto.String("[invalid-regex"),
		}.Build()

		_, err := util.ResolveSecret(ctx, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid validation regex")
	})
}
