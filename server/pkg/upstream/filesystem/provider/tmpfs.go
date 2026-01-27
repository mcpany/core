// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider provides access to a temporary in-memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider.
//
// Returns the result.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
// Returns the result.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
// virtualPath is the virtualPath.
//
// Returns the result.
// Returns an error if the operation fails.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// OpenFile opens a file using the Tmpfs filesystem, resolving the path first.
//
// path is the virtual path.
// flag is the open flag.
// perm is the permission.
//
// Returns the file.
// Returns an error if the operation fails.
func (p *TmpfsProvider) OpenFile(path string, flag int, perm os.FileMode) (afero.File, error) {
	resolvedPath, err := p.ResolvePath(path)
	if err != nil {
		return nil, err
	}
	return p.fs.OpenFile(resolvedPath, flag, perm)
}

// Close closes the provider.
//
// Returns an error if the operation fails.
func (p *TmpfsProvider) Close() error {
	return nil
}
