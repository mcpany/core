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

// NewTmpfsProvider creates a new tmpfs provider.
//
// Summary: Creates a new tmpfs provider.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *TmpfsProvider: The result.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs retrieves the fs.
//
// Summary: Retrieves the fs.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - afero.Fs: The result.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolvePath resolve path.
//
// Summary: ResolvePath resolve path.
//
// Parameters: - None.
//   - virtualPath (string): The virtual path.
//
// Returns: - None.
//   - string: The result.
//   - error: An error if the operation fails.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (p *TmpfsProvider) Close() error {
	return nil
}
