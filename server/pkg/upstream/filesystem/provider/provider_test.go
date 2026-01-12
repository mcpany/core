// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTmpfsProvider(t *testing.T) {
	p := NewTmpfsProvider()
	require.NotNil(t, p)
	fs := p.GetFs()
	require.NotNil(t, fs)

	path, err := p.ResolvePath("/some/path")
	assert.NoError(t, err)
	assert.Equal(t, "/some/path", path)

	err = p.Close()
	assert.NoError(t, err)
}

func TestZipProvider(t *testing.T) {
	// Create a dummy zip file
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("test.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("content"))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)

	tmpZip, err := os.CreateTemp("", "test*.zip")
	require.NoError(t, err)
	defer os.Remove(tmpZip.Name())

	_, err = tmpZip.Write(buf.Bytes())
	require.NoError(t, err)
	tmpZip.Close()

	pathStr := tmpZip.Name()
	config := &configv1.ZipFs{
		FilePath: &pathStr,
	}

	p, err := NewZipProvider(config)
	require.NoError(t, err)
	require.NotNil(t, p)

	fs := p.GetFs()
	require.NotNil(t, fs)

	path, err := p.ResolvePath("/test.txt")
	assert.NoError(t, err)
	// filepath.Clean depends on OS.
	assert.Equal(t, filepath.Clean("/test.txt"), path)

	err = p.Close()
	assert.NoError(t, err)

	// Test nil config
	_, err = NewZipProvider(nil)
	assert.Error(t, err)

	// Test invalid zip
	badZip, err := os.CreateTemp("", "bad*.zip")
	require.NoError(t, err)
	badZip.WriteString("not a zip")
	badZip.Close()
	defer os.Remove(badZip.Name())

	badPath := badZip.Name()
	_, err = NewZipProvider(&configv1.ZipFs{FilePath: &badPath})
	assert.Error(t, err)

	// Test missing file
	missingPath := "/missing/file.zip"
	_, err = NewZipProvider(&configv1.ZipFs{FilePath: &missingPath})
	assert.Error(t, err)
}

func TestS3Provider_ResolvePath(t *testing.T) {
	p := &S3Provider{fs: afero.NewMemMapFs()}

	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"test.txt", "test.txt", false},
		{"/test.txt", "test.txt", false},
		{"folder/test.txt", "folder/test.txt", false},
		{"/folder/test.txt", "folder/test.txt", false},
		{"../test.txt", "test.txt", false}, // Clean should handle ..
		{".", "", true},
		{"/", "", true},
	}

	for _, tt := range tests {
		got, err := p.ResolvePath(tt.input)
		if tt.wantErr {
			assert.Error(t, err, "input: %s", tt.input)
		} else {
			assert.NoError(t, err, "input: %s", tt.input)
			assert.Equal(t, tt.expected, got, "input: %s", tt.input)
		}
	}

	assert.NotNil(t, p.GetFs())
	assert.NoError(t, p.Close())
}

func TestGcsProvider_ResolvePath(t *testing.T) {
	p := &GcsProvider{fs: &gcsFs{}}

	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"test.txt", "test.txt", false},
		{"/test.txt", "test.txt", false},
		{"folder/test.txt", "folder/test.txt", false},
		{"/folder/test.txt", "folder/test.txt", false},
		{"../test.txt", "test.txt", false}, // Clean should handle ..
		{".", "", true},
		{"/", "", true},
	}

	for _, tt := range tests {
		got, err := p.ResolvePath(tt.input)
		if tt.wantErr {
			assert.Error(t, err, "input: %s", tt.input)
		} else {
			assert.NoError(t, err, "input: %s", tt.input)
			assert.Equal(t, tt.expected, got, "input: %s", tt.input)
		}
	}

	// GCS Close calls client.Close(), so we can't easily test it without a mock client or ensuring nil check.
	// gcs.go:
	// func (p *GcsProvider) Close() error {
	// 	if p.client != nil {
	// 		return p.client.Close()
	// 	}
	// 	return nil
	// }
	// So we can test Close() with nil client.
	assert.NotNil(t, p.GetFs())
	assert.NoError(t, p.Close())
}

func TestSftpProvider_ResolvePath(t *testing.T) {
	p := &SftpProvider{fs: &sftpFs{}}

	path, err := p.ResolvePath("/some/path")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Clean("/some/path"), path)

	// SFTP Close calls client.Close() and conn.Close().
	// func (p *SftpProvider) Close() error {
	// 	if p.client != nil {
	// 		_ = p.client.Close()
	// 	}
	// 	if p.conn != nil {
	// 		_ = p.conn.Close()
	// 	}
	// 	return nil
	// }
	assert.NotNil(t, p.GetFs())
	assert.NoError(t, p.Close())
}
