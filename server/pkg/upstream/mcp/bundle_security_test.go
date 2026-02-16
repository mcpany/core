// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMaliciousZip(t *testing.T, entries map[string]string) string {
	f, err := os.CreateTemp("", "malicious-*.zip")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	w := zip.NewWriter(f)
	defer func() { _ = w.Close() }()

	for name, content := range entries {
		// We use CreateHeader to bypass any potential cleaning in Create (though Create generally allows ..)
		header := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		writer, err := w.CreateHeader(header)
		require.NoError(t, err)
		_, err = io.WriteString(writer, content)
		require.NoError(t, err)
	}
	return f.Name()
}

func TestUnzipBundle_ZipSlip(t *testing.T) {
	tempDir := t.TempDir()

	// Create a zip with a file attempting to traverse up
	zipPath := createMaliciousZip(t, map[string]string{
		"../../evil.txt": "evil content",
	})
	defer func() { _ = os.Remove(zipPath) }()

	err := unzipBundle(zipPath, tempDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "illegal file path")

	// Verify the file was not created outside tempDir
	// The file would try to be created at tempDir/../../evil.txt
	// We check if it exists there.
	evilPath := filepath.Join(tempDir, "../../evil.txt")
	_, err = os.Stat(evilPath)
	assert.True(t, os.IsNotExist(err), "evil.txt should not exist at %s", evilPath)
}

func TestUnzipBundle_ZipBomb_FileSize(t *testing.T) {
	tempDir := t.TempDir()

	// Create a zip with a file that is small compressed but large uncompressed
	// We can simulate this by setting a small MaxFileSize.

	// Create a file with 100 bytes
	content := make([]byte, 100)
	for i := range content {
		content[i] = 'a'
	}

	zipPath := createMaliciousZip(t, map[string]string{
		"bomb.txt": string(content),
	})
	defer func() { _ = os.Remove(zipPath) }()

	// Set limit to 50 bytes
	err := unzipBundle(zipPath, tempDir, unzipOptions{
		MaxFileSize: 50,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestUnzipBundle_ZipBomb_TotalSize(t *testing.T) {
	tempDir := t.TempDir()

	// Create 3 files of 40 bytes each = 120 bytes total
	content := make([]byte, 40)
	for i := range content {
		content[i] = 'a'
	}

	zipPath := createMaliciousZip(t, map[string]string{
		"file1.txt": string(content),
		"file2.txt": string(content),
		"file3.txt": string(content),
	})
	defer func() { _ = os.Remove(zipPath) }()

	// Set total limit to 100 bytes
	err := unzipBundle(zipPath, tempDir, unzipOptions{
		MaxTotalSize: 100,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "total decompressed size exceeds limit")
}

func TestUnzipBundle_Symlink_Safety(t *testing.T) {
	// This test requires creating a zip with a symlink.
	// archive/zip doesn't support creating symlinks easily via Writer, we have to set external attributes.

	tempDir := t.TempDir()
	zipFile, err := os.CreateTemp("", "symlink-*.zip")
	require.NoError(t, err)
	defer func() { _ = os.Remove(zipFile.Name()) }()
	defer func() { _ = zipFile.Close() }()

	w := zip.NewWriter(zipFile)

	// Add a symlink entry
	// Symlink to /etc/passwd
	linkName := "passwd_link"
	linkTarget := "/etc/passwd"

	header := &zip.FileHeader{
		Name:   linkName,
		Method: zip.Deflate,
	}
	// Set symlink mode
	header.SetMode(os.ModeSymlink | 0755)

	writer, err := w.CreateHeader(header)
	require.NoError(t, err)
	_, err = writer.Write([]byte(linkTarget))
	require.NoError(t, err)

	// Add a regular file
	fileHeader := &zip.FileHeader{
		Name: "regular.txt",
		Method: zip.Deflate,
	}
	fileWriter, err := w.CreateHeader(fileHeader)
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("regular content"))
	require.NoError(t, err)

	require.NoError(t, w.Close())
	require.NoError(t, zipFile.Close())

	// Unzip
	// Currently unzipBundle creates regular files for everything, ignoring symlink mode,
	// because it does os.OpenFile with Create.
	// This effectively disables symlinks, which is SAFE but maybe not intended feature-wise.
	// But from security perspective, it's safe as long as we don't accidentally write to a location
	// assuming it's a file when it's a link, but we are creating new files in empty dir.

	err = unzipBundle(zipFile.Name(), tempDir)
	require.NoError(t, err)

	// Check if passwd_link is a file containing "/etc/passwd"
	linkPath := filepath.Join(tempDir, linkName)
	info, err := os.Lstat(linkPath)
	require.NoError(t, err)

	// It should be a regular file, NOT a symlink, based on current implementation
	assert.True(t, info.Mode().IsRegular())
	assert.False(t, info.Mode()&os.ModeSymlink != 0)

	content, err := os.ReadFile(linkPath)
	require.NoError(t, err)
	assert.Equal(t, "/etc/passwd", string(content))
}
