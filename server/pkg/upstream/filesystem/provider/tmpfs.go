// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
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
//
// Returns:
//   - *TmpfsProvider: The resulting instance.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
// Returns the result.
//
// Returns:
//   - afero.Fs: The result.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
// virtualPath is the virtualPath.
//
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//   - virtualPath: string. The virtualPath.
//
// Returns:
//   - string: The result.
//   - error: An error if the operation fails.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
//
// Returns an error if the operation fails.
//
// Returns:
//   - error: An error if the operation fails.
func (p *TmpfsProvider) Close() error {
	return nil
}
