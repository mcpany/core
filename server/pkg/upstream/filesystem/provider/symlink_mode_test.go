package provider

import (
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider_SymlinkMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root directory
	rootDir := filepath.Join(tmpDir, "root")
	err := os.Mkdir(rootDir, 0755)
	require.NoError(t, err)

	// Create a file inside root
	insideFile := filepath.Join(rootDir, "inside.txt")
	err = os.WriteFile(insideFile, []byte("inside"), 0644)
	require.NoError(t, err)

	// Create a file outside root
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("outside"), 0644)
	require.NoError(t, err)

	// Create an internal symlink
	internalLink := filepath.Join(rootDir, "link_inside")
	err = os.Symlink(insideFile, internalLink)
	require.NoError(t, err)

	// Create an external symlink
	externalLink := filepath.Join(rootDir, "link_outside")
	err = os.Symlink(outsideFile, externalLink)
	require.NoError(t, err)

	tests := []struct {
		name        string
		mode        configv1.FilesystemUpstreamService_SymlinkMode
		virtualPath string
		wantErr     bool
		errContains string
	}{
		{
			name:        "ALLOW mode - internal link",
			mode:        configv1.FilesystemUpstreamService_ALLOW,
			virtualPath: "/link_inside",
			wantErr:     false,
		},
		{
			name:        "ALLOW mode - external link",
			mode:        configv1.FilesystemUpstreamService_ALLOW,
			virtualPath: "/link_outside",
			wantErr:     true,
			errContains: "path traversal detected",
		},
		{
			name:        "DENY mode - internal link",
			mode:        configv1.FilesystemUpstreamService_DENY,
			virtualPath: "/link_inside",
			wantErr:     true,
			errContains: "symlinks are disabled",
		},
		{
			name:        "DENY mode - external link",
			mode:        configv1.FilesystemUpstreamService_DENY,
			virtualPath: "/link_outside",
			wantErr:     true,
			errContains: "symlinks are disabled",
		},
		{
			name:        "DENY mode - direct file",
			mode:        configv1.FilesystemUpstreamService_DENY,
			virtualPath: "/inside.txt",
			wantErr:     false,
		},
		{
			name:        "INTERNAL_ONLY mode - internal link",
			mode:        configv1.FilesystemUpstreamService_INTERNAL_ONLY,
			virtualPath: "/link_inside",
			wantErr:     false,
		},
		{
			name:        "INTERNAL_ONLY mode - external link",
			mode:        configv1.FilesystemUpstreamService_INTERNAL_ONLY,
			virtualPath: "/link_outside",
			wantErr:     true,
			errContains: "path traversal detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLocalProvider(nil, map[string]string{"/": rootDir}, nil, nil, tt.mode)
			_, err := p.ResolvePath(tt.virtualPath)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
