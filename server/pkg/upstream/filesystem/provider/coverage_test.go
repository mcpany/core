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
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// --- GcsFs Unit Tests ---
func TestGcsFs_Structure(t *testing.T) {
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
	tempDir := t.TempDir()

	// Allow tempDir for validation
	validation.SetAllowedPaths([]string{tempDir})
	defer validation.SetAllowedPaths(nil)

	tmpZip := filepath.Join(tempDir, "test.zip")

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
	cfg := configv1.ZipFs_builder{FilePath: proto.String(tmpZip)}.Build()
	p, err := NewZipProvider(cfg)
	require.NoError(t, err)
	defer p.Close()

	assert.NotNil(t, p.GetFs())

	res, err := p.ResolvePath("foo/../bar")
	assert.NoError(t, err)
	assert.Equal(t, "bar", res)
}

func TestZipProvider_Errors(t *testing.T) {
	// 1. File not found
	cfg1 := configv1.ZipFs_builder{FilePath: proto.String("non_existent.zip")}.Build()
	_, err := NewZipProvider(cfg1)
	assert.Error(t, err)

	// 2. Not a zip file
	tmpFile := filepath.Join(t.TempDir(), "not_a_zip.txt")
	err = os.WriteFile(tmpFile, []byte("not a zip"), 0644)
	require.NoError(t, err)

	cfg2 := configv1.ZipFs_builder{FilePath: proto.String(tmpFile)}.Build()
	_, err = NewZipProvider(cfg2)
	assert.Error(t, err)
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

// --- GcsFile Methods Error Checking ---
func TestGcsFile_Methods(t *testing.T) {
	// This covers the error paths when file is not properly opened
	f := &gcsFile{}

	_, err := f.Read([]byte{})
	assert.Error(t, err)

	_, err = f.Write([]byte{})
	assert.Error(t, err)

	_, err = f.Seek(0, 0)
	assert.Error(t, err)

	_, err = f.WriteAt(nil, 0)
	assert.Error(t, err)

	assert.NoError(t, f.Sync())
	assert.Error(t, f.Truncate(0))
	assert.NoError(t, f.Close())

	assert.Equal(t, "", f.Name())
}

func TestGcsFs_FileCreation(t *testing.T) {
	// We can't really test Create/Open without client, but we can verify the interface types
	fs := &gcsFs{bucket: "b", ctx: context.Background()}
	var _ afero.Fs = fs
}
