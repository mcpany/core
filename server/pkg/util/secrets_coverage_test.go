// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_CoverageExtra(t *testing.T) {
	t.Run("Regex Validation Compilation Failure", func(t *testing.T) {
		secret := &configv1.SecretValue{
            Value: &configv1.SecretValue_PlainText{
                PlainText: "secret_value",
            },
			ValidationRegex: "(?i)[a-z", // Invalid regex
		}

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid validation regex")
	})

	t.Run("Regex Validation Mismatch", func(t *testing.T) {
		secret := &configv1.SecretValue{
            Value: &configv1.SecretValue_PlainText{
                PlainText: "12345",
            },
			ValidationRegex: "^[a-z]+$", // Only lowercase letters
		}

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match validation regex")
	})

	t.Run("Max Recursion Depth Exceeded", func(t *testing.T) {
		secret := &configv1.SecretValue{
            Value: &configv1.SecretValue_PlainText{
                PlainText: "value",
            },
		}

		// Call resolveSecretImpl directly with a depth greater than maxSecretRecursionDepth
		_, err := resolveSecretImpl(context.Background(), secret, maxSecretRecursionDepth+1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeded max recursion depth")
	})

	t.Run("Missing Environment Variable with Restricted Name", func(t *testing.T) {
        // Test when IsEnvVarAllowed returns false
		secret := &configv1.SecretValue{
            Value: &configv1.SecretValue_EnvironmentVariable{
                EnvironmentVariable: "PATH", // Assuming PATH is restricted by IsEnvVarAllowed
            },
		}

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
        assert.Contains(t, err.Error(), "restricted")
	})

    t.Run("Invalid File Path Allowed List Error", func(t *testing.T) {
        validation.SetAllowedPaths([]string{"/allowed/path"})
        t.Cleanup(func() { validation.SetAllowedPaths(nil) })

        secret := &configv1.SecretValue{
            Value: &configv1.SecretValue_FilePath{
                FilePath: "/disallowed/path/secret.txt",
            },
        }

        _, err := ResolveSecret(context.Background(), secret)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid secret file path")
    })

    t.Run("File Read Error (Path is Directory)", func(t *testing.T) {
        tmpdir, err := os.MkdirTemp("", "test_secret_dir")
        assert.NoError(t, err)
        defer os.RemoveAll(tmpdir)

        validation.SetAllowedPaths([]string{tmpdir})
        t.Cleanup(func() { validation.SetAllowedPaths(nil) })

        secret := &configv1.SecretValue{
            Value: &configv1.SecretValue_FilePath{
                FilePath: tmpdir,
            },
        } // Passing a directory instead of a file should cause a read error

        _, err = ResolveSecret(context.Background(), secret)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to read secret from file")
    })
}
