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
// Parameters:
//   None.
//
// Returns:
//   - *TmpfsProvider: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs retrieves the fs.
//
// Summary: Retrieves the fs.
//
// Parameters:
//   None.
//
// Returns:
//   - afero.Fs: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolvePath resolve path.
//
// Summary: ResolvePath resolve path.
//
// Parameters:
//   - virtualPath (string): The virtual path.
//
// Returns:
//   - string: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
