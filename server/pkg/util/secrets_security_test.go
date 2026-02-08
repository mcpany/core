package util_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
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
	_, err := util.ResolveSecret(context.Background(), secret)
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
	// t.Setenv("MCPANY_FILE_PATH_ALLOW_LIST", tempDir)
	validation.SetAllowedPaths([]string{tempDir})
	defer func() {
		validation.SetAllowedPaths(nil)
		_ = os.RemoveAll(tempDir)
	}()

	secretFile := filepath.Join(tempDir, "my..secret.txt")
	err = os.WriteFile(secretFile, []byte("VALID_SECRET"), 0600)
	require.NoError(t, err)

	secret := &configv1.SecretValue{}
	secret.SetFilePath(secretFile)

	resolved, err := util.ResolveSecret(context.Background(), secret)
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

	_, err := util.ResolveSecret(context.Background(), secret)
	assert.Error(t, err)
	// Currently (before fix), this will error with network timeout or similar,
	// NOT "blocked link-local IP".
	// After fix, it MUST contain "blocked link-local IP".

	// We can assert that the error contains the specific message.
	assert.Contains(t, err.Error(), "resolved to link-local ip")
}

func TestResolveSecret_SSRF_PrivateIP_Blocked(t *testing.T) {
	// Attempt to access a private IP (e.g. 192.168.1.1)
	// This should be blocked by our SSRF protection (if enhanced).

	// Force block by setting env var to false (or ensuring default is block)
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_SECRETS", "false")

	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl("http://192.168.1.1/secret")
	secret := &configv1.SecretValue{}
	secret.SetRemoteContent(remoteContent)

	_, err := util.ResolveSecret(context.Background(), secret)
	assert.Error(t, err, "ResolveSecret should block SSRF to private IP")

	// Before fix, this might fail with "connection refused" or "i/o timeout" or similar.
	// After fix, it MUST contain "blocked private IP".
	assert.Contains(t, err.Error(), "resolved to private ip")
}

func TestResolveSecret_SSRF_PrivateIP_Allowed(t *testing.T) {
	// Attempt to access a private IP (e.g. 192.168.1.1)
	// This should be ALLOWED if env var is true.
	// Since we can't actually connect, it will timeout or fail with network error.

	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_SECRETS", "true")

	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl("http://192.168.1.1/secret")
	secret := &configv1.SecretValue{}
	secret.SetRemoteContent(remoteContent)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := util.ResolveSecret(ctx, secret)
	assert.Error(t, err)

	// It should NOT be "blocked private IP"
	assert.NotContains(t, err.Error(), "resolved to private ip")
}

func TestResolveSecret_AbsolutePathsBlockedByDefault(t *testing.T) {
	// Ensure env var is NOT set
	// os.Unsetenv("MCPANY_FILE_PATH_ALLOW_LIST")
	validation.SetAllowedPaths(nil)

	// Create a temp file
	// os.CreateTemp uses default temp dir, which is usually outside CWD.
	tmpFile, err := os.CreateTemp("", "secret_test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("supersecret")
	require.NoError(t, err)
	tmpFile.Close()

	absPath, err := filepath.Abs(tmpFile.Name())
	require.NoError(t, err)

	secret := &configv1.SecretValue{}
	secret.SetFilePath(absPath)

	// This should FAIL because /tmp is not in CWD and not in allow list
	_, err = util.ResolveSecret(context.Background(), secret)
	assert.Error(t, err, "ResolveSecret should block absolute paths outside CWD by default")
	if err != nil {
		assert.Contains(t, err.Error(), "path")
		assert.Contains(t, err.Error(), "is not allowed")
	}
}

func TestResolveSecret_AbsolutePathsAllowedWithList(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "secret_test_allowed")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Add the temp dir to the allow list
	// t.Setenv("MCPANY_FILE_PATH_ALLOW_LIST", filepath.Dir(tmpFile.Name()))
	validation.SetAllowedPaths([]string{filepath.Dir(tmpFile.Name())})
	defer validation.SetAllowedPaths(nil)

	_, err = tmpFile.WriteString("supersecret")
	require.NoError(t, err)
	tmpFile.Close()

	absPath, err := filepath.Abs(tmpFile.Name())
	require.NoError(t, err)

	secret := &configv1.SecretValue{}
	secret.SetFilePath(absPath)

	// This should pass
	val, err := util.ResolveSecret(context.Background(), secret)
	assert.NoError(t, err)
	assert.Equal(t, "supersecret", val)
}
