// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFilesystemUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestFilesystemUpstream_CreateFilesystem_Unsupported(t *testing.T) {
	u := &Upstream{}

	tests := []struct {
		name        string
		config      *configv1.FilesystemUpstreamService
		expectedErr string
	}{
		{
			name: "http",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Http{
					Http: &configv1.HttpFs{},
				},
			},
			expectedErr: "http filesystem is not yet supported",
		},
		{
			name: "zip",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Zip{
					Zip: &configv1.ZipFs{},
				},
			},
			expectedErr: "zip filesystem is not yet supported",
		},
		{
			name: "gcs",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Gcs{
					Gcs: &configv1.GcsFs{},
				},
			},
			expectedErr: "gcs filesystem is not yet supported",
		},
		{
			name: "sftp",
			config: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Sftp{
					Sftp: &configv1.SftpFs{},
				},
			},
			expectedErr: "sftp filesystem is not yet supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := u.createFilesystem(tt.config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestFilesystemUpstream_ResolvePath_EdgeCases(t *testing.T) {
	u := &Upstream{}

	// Create temp dir for OsFs testing
	tempDir, err := os.MkdirTemp("", "fs_resolve_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	realPath, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)

	osConfig := &configv1.FilesystemUpstreamService{
		RootPaths: map[string]string{
			"/data": realPath,
		},
		FilesystemType: &configv1.FilesystemUpstreamService_Os{
			Os: &configv1.OsFs{},
		},
	}

	t.Run("Tmpfs_Resolve", func(t *testing.T) {
		tmpConfig := &configv1.FilesystemUpstreamService{
			FilesystemType: &configv1.FilesystemUpstreamService_Tmpfs{
				Tmpfs: &configv1.MemMapFs{},
			},
		}
		path, err := u.resolvePath("/foo/bar", tmpConfig)
		assert.NoError(t, err)
		assert.Equal(t, "/foo/bar", path)
	})

	t.Run("OsFs_NoRootPaths", func(t *testing.T) {
		badConfig := &configv1.FilesystemUpstreamService{
			RootPaths: map[string]string{},
			FilesystemType: &configv1.FilesystemUpstreamService_Os{
				Os: &configv1.OsFs{},
			},
		}
		_, err := u.resolvePath("/data/foo", badConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no root paths defined")
	})

	t.Run("OsFs_NoMatchingRoot", func(t *testing.T) {
		_, err := u.resolvePath("/other/foo", osConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed (no matching root)")
	})

	t.Run("OsFs_InvalidRootPath", func(t *testing.T) {
		badRootConfig := &configv1.FilesystemUpstreamService{
			RootPaths: map[string]string{
				"/data": "/non/existent/path",
			},
			FilesystemType: &configv1.FilesystemUpstreamService_Os{
				Os: &configv1.OsFs{},
			},
		}
		_, err := u.resolvePath("/data/foo", badRootConfig)
		assert.Error(t, err)
		// It might fail at EvalSymlinks or Abs
		assert.Contains(t, err.Error(), "failed to resolve root path")
	})

	t.Run("OsFs_PathTraversal", func(t *testing.T) {
		// Attempt to go up out of root
		_, err := u.resolvePath("/data/../foo", osConfig)
		// resolvedPath would be /foo, which doesn't start with realPath
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("OsFs_SymlinkTraversal", func(t *testing.T) {
		// Create a symlink pointing outside
		secretFile := filepath.Join(os.TempDir(), "secret.txt")
		err := os.WriteFile(secretFile, []byte("secret"), 0644)
		require.NoError(t, err)
		defer os.Remove(secretFile)

		symlink := filepath.Join(tempDir, "link")
		err = os.Symlink(secretFile, symlink)
		require.NoError(t, err)

		_, err = u.resolvePath("/data/link", osConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("OsFs_NonExistentPath_ParentExists", func(t *testing.T) {
		// /data/newfile.txt -> should resolve correctly even if file doesn't exist, as long as parent exists and is safe
		path, err := u.resolvePath("/data/newfile.txt", osConfig)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(realPath, "newfile.txt"), path)
	})

	t.Run("OsFs_NonExistentPath_ParentDoesNotExist", func(t *testing.T) {
		// /data/nested/newfile.txt -> /data exists, /data/nested does not.
		// resolvePath currently handles non-existent files by checking ancestors.
		// If /data/nested doesn't exist, it checks /data.
		// /data is safe. So /data/nested/newfile.txt should be safe.
		path, err := u.resolvePath("/data/nested/newfile.txt", osConfig)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(realPath, "nested/newfile.txt"), path)
	})

	t.Run("OsFs_DefaultFallback", func(t *testing.T) {
		// Test fallback to default filesystem type (OsFs)
		config := &configv1.FilesystemUpstreamService{
			RootPaths: map[string]string{
				"/data": realPath,
			},
			// No FilesystemType set
		}
		path, err := u.resolvePath("/data/foo", config)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(realPath, "foo"), path)
	})
}

func TestFilesystemUpstream_Register_Error(t *testing.T) {
	u := NewUpstream()
	b, _ := bus.NewProvider(nil)
	tm := tool.NewManager(b)

	// Nil config
	_, _, _, err := u.Register(context.Background(), nil, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Equal(t, "service config is nil", err.Error())

	// Empty name
	configEmptyName := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{},
		},
	}
	_, _, _, err = u.Register(context.Background(), configEmptyName, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Equal(t, "id cannot be empty", err.Error())

	// No filesystem config
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: nil,
		},
	}
	_, _, _, err = u.Register(context.Background(), config, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filesystem service config is nil")

	// Invalid filesystem creation (e.g. http unsupported)
	configHttp := &configv1.UpstreamServiceConfig{
		Name: proto.String("test_fs_http"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Http{
					Http: &configv1.HttpFs{},
				},
			},
		},
	}
	_, _, _, err = u.Register(context.Background(), configHttp, tm, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http filesystem is not yet supported")
}
