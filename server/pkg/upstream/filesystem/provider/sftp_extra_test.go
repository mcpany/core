package provider

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSftpProvider_Extended(t *testing.T) {
	// Create a temporary directory for the SFTP server to serve
	addr, _, cleanup := startSFTPServer(t, nil)
	defer cleanup()

	// Create temp dir for testing
	tmpDir, err := os.MkdirTemp("", "sftp-test-extended")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := configv1.SftpFs_builder{
		Address:  proto.String(addr),
		Username: proto.String("testuser"),
		Password: proto.String("testpass"),
	}.Build()

	provider, err := NewSftpProvider(config)
	require.NoError(t, err)
	defer provider.Close()

	fs := provider.GetFs()

	t.Run("Mkdir and MkdirAll", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "subdir")
		err := fs.Mkdir(dir, 0755)
		require.NoError(t, err)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		nested := filepath.Join(tmpDir, "nested/a/b")
		err = fs.MkdirAll(nested, 0755)
		require.NoError(t, err)

		info, err = os.Stat(nested)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("Remove and RemoveAll", func(t *testing.T) {
		// Create file to remove
		f := filepath.Join(tmpDir, "remove.txt")
		err := os.WriteFile(f, []byte("test"), 0644)
		require.NoError(t, err)

		err = fs.Remove(f)
		require.NoError(t, err)
		_, err = os.Stat(f)
		assert.True(t, os.IsNotExist(err))

		// Create dir to remove all
		d := filepath.Join(tmpDir, "removeall")
		err = os.MkdirAll(d, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(d, "file.txt"), []byte("test"), 0644)
		require.NoError(t, err)

		err = fs.RemoveAll(d)
		require.NoError(t, err)
		_, err = os.Stat(d)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Rename", func(t *testing.T) {
		oldPath := filepath.Join(tmpDir, "old.txt")
		newPath := filepath.Join(tmpDir, "new.txt")
		err := os.WriteFile(oldPath, []byte("test"), 0644)
		require.NoError(t, err)

		err = fs.Rename(oldPath, newPath)
		require.NoError(t, err)

		_, err = os.Stat(oldPath)
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat(newPath)
		assert.NoError(t, err)
	})

	t.Run("Chmod and Chown and Chtimes", func(t *testing.T) {
		f := filepath.Join(tmpDir, "meta.txt")
		err := os.WriteFile(f, []byte("test"), 0644)
		require.NoError(t, err)

		err = fs.Chmod(f, 0600)
		require.NoError(t, err)

		now := time.Now()
		err = fs.Chtimes(f, now, now)
		require.NoError(t, err)
	})

	t.Run("OpenFile", func(t *testing.T) {
		f := filepath.Join(tmpDir, "openfile.txt")
		// Write mode
		file, err := fs.OpenFile(f, os.O_CREATE|os.O_WRONLY, 0644)
		require.NoError(t, err)
		_, err = file.WriteString("content")
		require.NoError(t, err)
		file.Close()

		// Read mode
		file, err = fs.OpenFile(f, os.O_RDONLY, 0)
		require.NoError(t, err)
		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))
		file.Close()
	})

	t.Run("ReadDir", func(t *testing.T) {
		d := filepath.Join(tmpDir, "readdir")
		err := os.Mkdir(d, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(d, "a.txt"), []byte("a"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(d, "b.txt"), []byte("b"), 0644)
		require.NoError(t, err)

		f, err := fs.Open(d)
		require.NoError(t, err)
		defer f.Close()

		names, err := f.Readdirnames(-1)
		require.NoError(t, err)
		assert.Contains(t, names, "a.txt")
		assert.Contains(t, names, "b.txt")
	})
}
