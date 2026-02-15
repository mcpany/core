package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewZipProvider_FileNotFound(t *testing.T) {
	// Use a path in the current directory to pass IsAllowedPath check
	path := "non-existent-file.zip"
	config := configv1.ZipFs_builder{
		FilePath: proto.String(path),
	}.Build()
	p, err := NewZipProvider(config)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "failed to open zip file")
}

func TestNewSftpProvider_InvalidKeyFile(t *testing.T) {
	config := configv1.SftpFs_builder{
		Address:  proto.String("127.0.0.1:2222"),
		Username: proto.String("user"),
		KeyPath:  proto.String("/path/to/non/existent/key"),
	}.Build()
	p, err := NewSftpProvider(config)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "failed to read private key")
}

func TestSftpProvider_ConfigValidation(t *testing.T) {
	// Test address formatting logic (implicit port 22)
	// We can't easily inspect the internal addr variable in NewSftpProvider,
	// but we can verify that it proceeds to Dial (and fails) rather than failing earlier.

	// However, we can test the Key parsing failure which happens before Dial but after reading file.
	// Create a temporary invalid key file
	tmpFile, err := os.CreateTemp("", "invalid_key")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid key content")
	require.NoError(t, err)
	tmpFile.Close()

	config := configv1.SftpFs_builder{
		Address:  proto.String("127.0.0.1"), // No port, should add :22
		Username: proto.String("user"),
		KeyPath:  proto.String(tmpFile.Name()),
	}.Build()

	p, err := NewSftpProvider(config)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestLocalProvider_ResolveNonExistentPath_EdgeCases(t *testing.T) {
	// Setup a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "local_provider_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a real directory
	realDir := filepath.Join(tmpDir, "real")
	err = os.Mkdir(realDir, 0755)
	require.NoError(t, err)

	// Create a symlink to the real directory
	linkDir := filepath.Join(tmpDir, "link")
	err = os.Symlink(realDir, linkDir)
	require.NoError(t, err)

	// Create a broken symlink
	brokenLink := filepath.Join(realDir, "broken")
	err = os.Symlink("/non/existent/target", brokenLink)
	require.NoError(t, err)

	p := NewLocalProvider(nil, map[string]string{
		"/": tmpDir,
	}, nil, nil, 0)

	// Case 1: Resolve a path that goes through a valid symlink to a non-existent file
	path1 := "/link/non_existent_file.txt"
	resolved, err := p.ResolvePath(path1)
	assert.NoError(t, err)
	// The resolved path should be in the real directory
	expected := filepath.Join(realDir, "non_existent_file.txt")
	// On some systems /tmp is a symlink, so EvalSymlinks on expected might be needed
	expectedCanonical, _ := filepath.EvalSymlinks(realDir)
	expected = filepath.Join(expectedCanonical, "non_existent_file.txt")
	assert.Equal(t, expected, resolved)

	// Case 2: Resolve a path that goes through a broken symlink (should fail)
	path2 := "/real/broken/file.txt"
	_, err = p.ResolvePath(path2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied: component broken is a broken symlink")
}

func TestNewGcsProvider_NilConfig(t *testing.T) {
	_, err := NewGcsProvider(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gcs config is nil")
}

func TestNewS3Provider_NilConfig(t *testing.T) {
	// NewS3Provider doesn't check for nil config explicitly at start?
	// Let's check s3.go source again.
	// It says:
	// func NewS3Provider(config *configv1.S3Fs) (*S3Provider, error) {
	// 	if config == nil {
	// 		return nil, fmt.Errorf("s3 config is nil")
	// 	}
	// ...
	_, err := NewS3Provider(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "s3 config is nil")
}
