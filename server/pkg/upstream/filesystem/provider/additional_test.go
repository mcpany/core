// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string {
	return &s
}

func TestLocalProvider_Methods(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewLocalProvider(nil, map[string]string{"/": tmpDir})

	// Test GetFs
	assert.NotNil(t, p.GetFs())

	// Test Close
	assert.NoError(t, p.Close())
}

func TestNewS3Provider_Config(t *testing.T) {
	// Minimal valid config
	cfg := &configv1.S3Fs{
		Bucket: stringPtr("my-bucket"),
		Region: stringPtr("us-east-1"),
	}
	p, err := NewS3Provider(cfg)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.NotNil(t, p.GetFs())
	assert.NoError(t, p.Close())

	// Config with credentials
	cfgWithCreds := &configv1.S3Fs{
		Bucket:          stringPtr("my-bucket"),
		Region:          stringPtr("us-east-1"),
		AccessKeyId:     stringPtr("AKIA..."),
		SecretAccessKey: stringPtr("test-secret-key"),
		SessionToken:    stringPtr("token"),
		Endpoint:        stringPtr("http://localhost:9000"),
	}
	p2, err := NewS3Provider(cfgWithCreds)
	require.NoError(t, err)
	require.NotNil(t, p2)
}

func TestNewGcsProvider_Error(t *testing.T) {
	// Nil config
	p, err := NewGcsProvider(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "gcs config is nil")
}

func TestNewSftpProvider_Error(t *testing.T) {
	// Nil config
	p, err := NewSftpProvider(nil)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Contains(t, err.Error(), "sftp config is nil")

	// Invalid key path
	cfg := &configv1.SftpFs{
		KeyPath: stringPtr("/non/existent/key"),
	}
	p, err = NewSftpProvider(cfg)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestNewZipProvider_Error(t *testing.T) {
	// Non-existent file
	path := "/non/existent/file.zip"
	cfg := &configv1.ZipFs{
		FilePath: &path,
	}
	p, err := NewZipProvider(cfg)
	assert.Error(t, err)
	assert.Nil(t, p)

	// Create a non-zip file
	tmpFile, err := os.CreateTemp("", "not-a-zip")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("not a zip"))
	tmpFile.Close()

	path2 := tmpFile.Name()
	cfg2 := &configv1.ZipFs{
		FilePath: &path2,
	}
	p2, err := NewZipProvider(cfg2)
	assert.Error(t, err)
	assert.Nil(t, p2)
}
