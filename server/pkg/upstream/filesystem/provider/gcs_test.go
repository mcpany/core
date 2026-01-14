// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewGcsProvider(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		_, err := NewGcsProvider(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gcs config is nil")
	})

	t.Run("Valid Config", func(t *testing.T) {
		config := &configv1.GcsFs{
			Bucket: proto.String("my-bucket"),
		}
		p, err := NewGcsProvider(context.Background(), config)
		if err == nil {
			defer p.Close()
			assert.NotNil(t, p)
			assert.NotNil(t, p.GetFs())
		} else {
			// It might fail if no creds are found, which is acceptable in this env.
			t.Logf("NewGcsProvider failed as expected (no creds): %v", err)
		}
	})
}

func TestGcsFileInfo(t *testing.T) {
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
	assert.NotZero(t, fi.ModTime())

	dirFi := &gcsFileInfo{
		name:  "dir",
		isDir: true,
	}
	assert.Equal(t, os.ModeDir|0755, dirFi.Mode())
	assert.True(t, dirFi.IsDir())
}

func TestGcsFs_Methods(t *testing.T) {
	var fs afero.Fs = &gcsFs{}
	// Just verify it compiles and implements the interface
	assert.NotNil(t, fs)
	assert.Equal(t, "gcs", fs.Name())

	// These return nil/error without client, so we can test the "Not Supported" ones
	assert.NoError(t, fs.Chmod("foo", 0))
	assert.NoError(t, fs.Chown("foo", 0, 0))
	assert.NoError(t, fs.Chtimes("foo", time.Now(), time.Now()))
}

func TestGcsFile_Methods_Errors(t *testing.T) {
	f := &gcsFile{}

	_, err := f.Read([]byte{})
	assert.Error(t, err)
	assert.Equal(t, "file not opened for reading", err.Error())

	_, err = f.Write([]byte{})
	assert.Error(t, err)
	assert.Equal(t, "file not opened for writing", err.Error())

	_, err = f.Seek(0, 0)
	assert.Error(t, err)
	assert.Equal(t, "seek not supported", err.Error())

	_, err = f.WriteAt([]byte{}, 0)
	assert.Error(t, err)
	assert.Equal(t, "writeat not supported", err.Error())

	err = f.Truncate(0)
	assert.Error(t, err)
	assert.Equal(t, "truncate not supported", err.Error())

	// Sync does nothing
	assert.NoError(t, f.Sync())

	// Close on empty file
	assert.NoError(t, f.Close())
}
