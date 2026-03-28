// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider provides access to a temporary in-memory filesystem.
//
// Summary: Represents a TmpfsProvider.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider provides newtmpfsprovider functionality.
//
// Summary: NewTmpfsProvider.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TmpfsProvider.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs provides getfs functionality.
//
// Summary: GetFs.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath provides resolvepath functionality.
//
// Summary: ResolvePath.
//
// Parameters.
//   - virtualPath: The parameter.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - virtualPath: string.
//
// Returns:
//   - string.
//   - error.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - None.
//
// Returns:
//   - error.
func (p *TmpfsProvider) Close() error {
	return nil
}
