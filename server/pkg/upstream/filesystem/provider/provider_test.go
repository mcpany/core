package provider

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

	// Allow temp dir for validation
	validation.SetAllowedPaths([]string{filepath.Dir(tmpZip.Name())})
	defer validation.SetAllowedPaths(nil)

	_, err = tmpZip.Write(buf.Bytes())
	require.NoError(t, err)
	tmpZip.Close()

	pathStr := tmpZip.Name()
	config := configv1.ZipFs_builder{
		FilePath: proto.String(pathStr),
	}.Build()

	p, err := NewZipProvider(config)
	require.NoError(t, err)
	require.NotNil(t, p)

	fs := p.GetFs()
	require.NotNil(t, fs)

	path, err := p.ResolvePath("/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Clean("/test.txt"), path)

	err = p.Close()
	assert.NoError(t, err)
}

func TestSftpProvider_ResolvePath(t *testing.T) {
	p := &SftpProvider{fs: &sftpFs{}}

	path, err := p.ResolvePath("/some/path")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Clean("/some/path"), path)

	assert.NotNil(t, p.GetFs())
	assert.NoError(t, p.Close())
}
