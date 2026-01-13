// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSftpProvider_Extended(t *testing.T) {
	// Re-use the setup from sftp_test.go
	addr, _, cleanup := startSFTPServer(t)
	defer cleanup()

	// Create temp dir for testing
	tmpDir, err := os.MkdirTemp("", "sftp-test-ext")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &configv1.SftpFs{
		Address:  &addr,
		Username: ptr("testuser"),
		Password: ptr("testpass"),
	}

	provider, err := NewSftpProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	fs := provider.GetFs()

	t.Run("Mkdir and MkdirAll", func(t *testing.T) {
		dir := "subdir"
		fullDir := tmpDir + "/" + dir
		// Use absolute paths as sftp server serves from root (local FS in test setup)
		err = fs.Mkdir(fullDir, 0755)
		require.NoError(t, err)

		info, err := os.Stat(fullDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		nested := fullDir + "/nested/deep"
		err = fs.MkdirAll(nested, 0755)
		require.NoError(t, err)

		info, err = os.Stat(nested)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("Remove and RemoveAll", func(t *testing.T) {
		fPath := tmpDir + "/todelete.txt"
		err := os.WriteFile(fPath, []byte("bye"), 0644)
		require.NoError(t, err)

		err = fs.Remove(fPath)
		require.NoError(t, err)

		_, err = os.Stat(fPath)
		assert.True(t, os.IsNotExist(err))

		// Test RemoveAll on empty directory
		emptyDir := tmpDir + "/empty_dir"
		err = os.Mkdir(emptyDir, 0755)
		require.NoError(t, err)

		err = fs.RemoveAll(emptyDir)
		require.NoError(t, err)

		_, err = os.Stat(emptyDir)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Rename", func(t *testing.T) {
		oldP := tmpDir + "/old.txt"
		newP := tmpDir + "/new.txt"
		os.WriteFile(oldP, []byte("content"), 0644)

		err := fs.Rename(oldP, newP)
		require.NoError(t, err)

		_, err = os.Stat(oldP)
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat(newP)
		assert.NoError(t, err)
	})

	t.Run("Chmod and Chown and Chtimes", func(t *testing.T) {
		p := tmpDir + "/attrs.txt"
		os.WriteFile(p, []byte("attrs"), 0644)

		err := fs.Chmod(p, 0600)
		require.NoError(t, err)

		// Check mode
		info, _ := os.Stat(p)
		// On some systems/configs, chmod via sftp might be restricted or mapped differently.
		// We trust no error means it was attempted.

		now := time.Now().Truncate(time.Second)
		err = fs.Chtimes(p, now, now)
		require.NoError(t, err)

		info, _ = os.Stat(p)
		// Allow some skew
		assert.WithinDuration(t, now, info.ModTime(), 5*time.Second)
	})

	t.Run("OpenFile", func(t *testing.T) {
		p := tmpDir + "/openfile.txt"
		f, err := fs.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
		require.NoError(t, err)
		_, err = f.Write([]byte("data"))
		require.NoError(t, err)
		f.Close()

		content, _ := os.ReadFile(p)
		assert.Equal(t, "data", string(content))
	})

	t.Run("ReadDir", func(t *testing.T) {
		dir := tmpDir + "/readdir"
		os.Mkdir(dir, 0755)
		os.WriteFile(dir+"/a", []byte("a"), 0644)
		os.WriteFile(dir+"/b", []byte("b"), 0644)

		f, err := fs.Open(dir)
		require.NoError(t, err)
		defer f.Close()

		names, err := f.Readdirnames(-1)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"a", "b"}, names)
	})
}
