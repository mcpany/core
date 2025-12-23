// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_EnvAllowList(t *testing.T) {
	// Setup env vars to be accessed
	t.Setenv("SAFE_VAR", "safe_value")
	t.Setenv("SENSITIVE_VAR", "sensitive_value")

	t.Run("AllowList Unset - Default Allow All", func(t *testing.T) {
		// Ensure allow list is unset (assuming clean env or previous tests didn't leak)
		// Since we can't easily unset in parallel tests without side effects, we rely on t.Setenv isolation.
		// However, to be sure, we can't explicitly unset with t.Setenv.
		// We just assume it's not set.

		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("SENSITIVE_VAR")
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "sensitive_value", resolved)
	})

	t.Run("AllowList Set - Restrict", func(t *testing.T) {
		t.Setenv("MCPANY_ENV_VAR_ALLOW_LIST", "SAFE_VAR,ANOTHER_VAR")

		// Allowed var
		secretSafe := &configv1.SecretValue{}
		secretSafe.SetEnvironmentVariable("SAFE_VAR")
		resolved, err := util.ResolveSecret(context.Background(), secretSafe)
		assert.NoError(t, err)
		assert.Equal(t, "safe_value", resolved)

		// Denied var
		secretSensitive := &configv1.SecretValue{}
		secretSensitive.SetEnvironmentVariable("SENSITIVE_VAR")
		_, err = util.ResolveSecret(context.Background(), secretSensitive)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access to environment variable \"SENSITIVE_VAR\" is not allowed")
	})

	t.Run("AllowList Set Empty - Deny All", func(t *testing.T) {
		t.Setenv("MCPANY_ENV_VAR_ALLOW_LIST", "")

		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("SAFE_VAR")
		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access to environment variable \"SAFE_VAR\" is not allowed")
	})

	t.Run("AllowList With Spaces", func(t *testing.T) {
		t.Setenv("MCPANY_ENV_VAR_ALLOW_LIST", "SAFE_VAR, SENSITIVE_VAR ")

		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("SENSITIVE_VAR")
		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "sensitive_value", resolved)
	})
}
