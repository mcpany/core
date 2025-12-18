// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSecret_PathTraversal(t *testing.T) {
	// Construct a path that attempts to traverse up.
	// Since we don't know the CWD, we just use a path starting with ../
	// We don't care if the file exists, we expect the validation to fail BEFORE reading.

	traversalPath := filepath.Join("..", "secret.txt")
	// This results in "../secret.txt"

	secret := &configv1.SecretValue{}
	secret.SetFilePath(traversalPath)

	// This should now FAIL because of IsSecurePath check
	_, err := util.ResolveSecret(secret)
	assert.Error(t, err, "ResolveSecret should block traversal paths")
	if err != nil {
		assert.Contains(t, err.Error(), "invalid secret file path")
		assert.Contains(t, err.Error(), "path contains '..'")
	}
}

func TestResolveSecret_ValidPathWithDoubleDotsInName(t *testing.T) {
    // This test ensures we didn't break valid filenames like "my..secret.txt"
    tempDir, err := os.MkdirTemp("", "mcpany-repro-valid")
    require.NoError(t, err)
    defer func() { _ = os.RemoveAll(tempDir) }()

    secretFile := filepath.Join(tempDir, "my..secret.txt")
    err = os.WriteFile(secretFile, []byte("VALID_SECRET"), 0600)
    require.NoError(t, err)

    secret := &configv1.SecretValue{}
    secret.SetFilePath(secretFile)

    resolved, err := util.ResolveSecret(secret)
    assert.NoError(t, err)
    assert.Equal(t, "VALID_SECRET", resolved)
}

func TestResolveSecret_SSRF_Blocked(t *testing.T) {
	// Attempt to access AWS Metadata service IP
	// This should be blocked by our SSRF protection
	// The test expects the function to return an error explicitly stating it is blocked.
	// If it returns a network error (timeout/unreachable), it means protection is NOT active.

	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl("http://169.254.169.254/latest/meta-data/")
	secret := &configv1.SecretValue{}
	secret.SetRemoteContent(remoteContent)

	_, err := util.ResolveSecret(secret)
	assert.Error(t, err)
	// Currently (before fix), this will error with network timeout or similar,
	// NOT "blocked link-local IP".
	// After fix, it MUST contain "blocked link-local IP".

	// We can assert that the error contains the specific message.
	assert.Contains(t, err.Error(), "blocked link-local IP")
}
