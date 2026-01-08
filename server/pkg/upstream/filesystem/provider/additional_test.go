// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Test S3 Provider Factory
func TestNewS3Provider(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		cfg := &configv1.S3Fs{
			Bucket:          proto.String("my-bucket"),
			Region:          proto.String("us-west-2"),
			AccessKeyId:     proto.String("key"),
			SecretAccessKey: proto.String("secret"),
			Endpoint:        proto.String("http://localhost:9000"),
		}
		p, err := NewS3Provider(cfg)
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.NotNil(t, p.GetFs())
	})

	t.Run("Missing Bucket", func(t *testing.T) {
		cfg := &configv1.S3Fs{
			Region: proto.String("us-east-1"),
		}
		p, err := NewS3Provider(cfg)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("Region Only", func(t *testing.T) {
		cfg := &configv1.S3Fs{
			Bucket: proto.String("bkt"),
			Region: proto.String("us-east-1"),
		}
		p, err := NewS3Provider(cfg)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}

// Test SFTP Provider Factory
func TestNewSftpProvider(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		p, err := NewSftpProvider(nil)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, "sftp config is nil", err.Error())
	})

	t.Run("Key Path Read Error", func(t *testing.T) {
		cfg := &configv1.SftpFs{
			Address: proto.String("localhost:22"),
			KeyPath: proto.String("/non/existent/key"),
		}
		p, err := NewSftpProvider(cfg)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "failed to read private key")
	})

	t.Run("Invalid Private Key", func(t *testing.T) {
		// Create a dummy invalid key file
		tmpFile := filepath.Join(t.TempDir(), "invalid_key")
		err := os.WriteFile(tmpFile, []byte("not a key"), 0600)
		require.NoError(t, err)

		cfg := &configv1.SftpFs{
			Address: proto.String("localhost:22"),
			KeyPath: proto.String(tmpFile),
		}
		p, err := NewSftpProvider(cfg)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("Connection Error", func(t *testing.T) {
		// Try to connect to a closed port
		cfg := &configv1.SftpFs{
			Address:  proto.String("localhost:54321"), // Random port likely closed
			Username: proto.String("user"),
			Password: proto.String("pass"),
		}
		p, err := NewSftpProvider(cfg)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "failed to dial ssh")
	})
}

// Test Zip Provider Factory
func TestNewZipProvider_Error(t *testing.T) {
	t.Run("File Not Found", func(t *testing.T) {
		cfg := &configv1.ZipFs{
			FilePath: proto.String("/non/existent/file.zip"),
		}
		p, err := NewZipProvider(cfg)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "failed to open zip file")
	})

	t.Run("Not A Zip File", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "not_zip.txt")
		err := os.WriteFile(tmpFile, []byte("plain text"), 0600)
		require.NoError(t, err)

		cfg := &configv1.ZipFs{
			FilePath: proto.String(tmpFile),
		}
		p, err := NewZipProvider(cfg)
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "failed to create zip reader")
	})
}

func TestLocalProvider_NoRoots(t *testing.T) {
	p := NewLocalProvider(nil, nil)
	_, err := p.ResolvePath("/test")
	assert.Error(t, err)
	assert.Equal(t, "no root paths defined", err.Error())
}
