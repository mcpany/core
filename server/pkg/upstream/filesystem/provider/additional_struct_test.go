// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// Tests for gcsFs

func TestGcsFs_Structs(t *testing.T) {
	// gcsFs implements afero.Fs
	var fs afero.Fs = &gcsFs{}
	assert.Equal(t, "gcs", fs.Name())
	assert.NoError(t, fs.Mkdir("foo", 0755))
	assert.NoError(t, fs.MkdirAll("foo/bar", 0755))
	assert.NoError(t, fs.Chmod("foo", 0644))
	assert.NoError(t, fs.Chown("foo", 1000, 1000))
	assert.NoError(t, fs.Chtimes("foo", time.Now(), time.Now()))

	fi := &gcsFileInfo{
		name:    "test.txt",
		size:    123,
		modTime: time.Now(),
		isDir:   false,
	}
	assert.Equal(t, "test.txt", fi.Name())
	assert.Equal(t, int64(123), fi.Size())
	assert.Equal(t, os.FileMode(0644), fi.Mode())
	assert.False(t, fi.IsDir())
	assert.Nil(t, fi.Sys())

	dirFi := &gcsFileInfo{
		name:  "dir",
		isDir: true,
	}
	assert.True(t, dirFi.IsDir())
	assert.Equal(t, os.ModeDir|0755, dirFi.Mode())
}

func TestGcsFile_Structs(t *testing.T) {
	f := &gcsFile{
		name: "test.txt",
	}
	assert.Equal(t, "test.txt", f.Name())

	// Test error returns for unsupported methods
	_, err := f.Seek(0, 0)
	assert.Error(t, err)
	assert.Equal(t, "seek not supported", err.Error())

	_, err = f.WriteAt(nil, 0)
	assert.Error(t, err)
	assert.Equal(t, "writeat not supported", err.Error())

	err = f.Truncate(0)
	assert.Error(t, err)
	assert.Equal(t, "truncate not supported", err.Error())

	assert.NoError(t, f.Sync())

	// Test Read/Write checks
	_, err = f.Read(nil)
	assert.Error(t, err)
	assert.Equal(t, "file not opened for reading", err.Error())

	_, err = f.Write(nil)
	assert.Error(t, err)
	assert.Equal(t, "file not opened for writing", err.Error())
}

// Tests for sftpFs
func TestSftpFs_Structs(t *testing.T) {
	// sftpFs implements afero.Fs
	var fs afero.Fs = &sftpFs{}
	assert.Equal(t, "sftp", fs.Name())
}
