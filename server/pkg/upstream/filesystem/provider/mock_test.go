// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper Mocks ---

// MockFileInfo implements os.FileInfo
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *MockFileInfo) ModTime() time.Time { return m.modTime }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }

// --- GcsFs Unit Tests ---
func TestGcsFs_Methods(t *testing.T) {
	// We cannot create a functional gcsFs without a real storage.Client,
	// because Open/Create etc call client methods directly.
	// However, we can test the methods that don't use the client, or fail gracefully.

	fs := &gcsFs{
		client: nil, // Will panic if used? Yes.
		bucket: "test-bucket",
		ctx:    context.Background(),
	}

	// Name
	assert.Equal(t, "gcs", fs.Name())

	// Mkdir / MkdirAll (No-ops)
	assert.NoError(t, fs.Mkdir("foo", 0755))
	assert.NoError(t, fs.MkdirAll("foo/bar", 0755))

	// Chmod / Chown / Chtimes (Not supported)
	assert.NoError(t, fs.Chmod("foo", 0755))
	assert.NoError(t, fs.Chown("foo", 1000, 1000))
	assert.NoError(t, fs.Chtimes("foo", time.Now(), time.Now()))

	// ResolvePath (from GcsProvider)
	p := &GcsProvider{fs: fs}
	resolved, err := p.ResolvePath("foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", resolved)

	resolved, err = p.ResolvePath("/foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", resolved)

	_, err = p.ResolvePath(".")
	assert.Error(t, err)

	// Close (nil client)
	assert.NoError(t, p.Close())
}

// --- GcsFileInfo Tests ---
func TestGcsFileInfo(t *testing.T) {
	now := time.Now()
	fi := &gcsFileInfo{
		name:    "test.txt",
		size:    123,
		modTime: now,
		isDir:   false,
	}

	assert.Equal(t, "test.txt", fi.Name())
	assert.Equal(t, int64(123), fi.Size())
	assert.Equal(t, now, fi.ModTime())
	assert.False(t, fi.IsDir())
	assert.Equal(t, os.FileMode(0644), fi.Mode())
	assert.Nil(t, fi.Sys())

	fiDir := &gcsFileInfo{
		name:  "dir",
		isDir: true,
	}
	assert.True(t, fiDir.IsDir())
	assert.Equal(t, os.ModeDir|0755, fiDir.Mode())
}

// --- SftpProvider Extra Tests ---
// Since SftpProvider uses a private sftpFs struct, and we can't easily mock the SFTP client...
// But we can check ResolvePath and other non-network methods.

func TestSftpProvider_Methods(t *testing.T) {
	p := &SftpProvider{}
	assert.NoError(t, p.Close()) // nil closer

	path, err := p.ResolvePath("foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", path) // It just cleans the path

	// Test GetFs
	assert.Nil(t, p.GetFs())
}

// --- ZipProvider Extra Tests ---
func TestZipProvider_Methods(t *testing.T) {
	// Create a real zip file for testing
	tmpZip := filepath.Join(t.TempDir(), "test.zip")

	// Create zip
	zFile, err := os.Create(tmpZip)
	require.NoError(t, err)

	zw := zip.NewWriter(zFile)
	f, err := zw.Create("hello.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("world"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	require.NoError(t, zFile.Close())

	// Now load it
	cfg := &configv1.ZipFs{FilePath: &tmpZip}
	p, err := NewZipProvider(cfg)
	require.NoError(t, err)
	defer p.Close()

	assert.NotNil(t, p.GetFs())

	res, err := p.ResolvePath("foo/../bar")
	assert.NoError(t, err)
	assert.Equal(t, "bar", res)
}

// --- TmpfsProvider Extra Tests ---
func TestTmpfsProvider_Methods(t *testing.T) {
	p := NewTmpfsProvider()
	assert.NotNil(t, p.GetFs())

	path, err := p.ResolvePath("foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", path)

	assert.NoError(t, p.Close())
}

// --- LocalProvider Extra Coverage ---
func TestLocalProvider_Close(t *testing.T) {
	p := &LocalProvider{}
	assert.NoError(t, p.Close())
	assert.Nil(t, p.GetFs())
}
