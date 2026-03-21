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

// NewTmpfsProvider creates a new TmpfsProvider.
//
// Returns:
//   - *TmpfsProvider: The result.
//
// Side Effects:
//   - None.
//
// Summary: Initializes NewTmpfsProvider operation.
//
// Returns:
//   - *TmpfsProvider: The new tmpfs provider instance.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
// Returns:
//   - afero.Fs: The result.
//
// Side Effects:
//   - None.
//
// Summary: Retrieves GetFs operation.
//
// Returns:
//   - afero.Fs: The underlying afero filesystem.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
// Parameters:
//   - virtualPath (string): The parameter.
//
// Returns:
//   - string: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes ResolvePath operation.
//
// Parameters:
//   - virtualPath (string): The virtual path to resolve.
//
// Returns:
//   - string: The resolved local path.
//   - error: An error if resolution fails.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Executes Close operation.
//
// Returns:
//   - error: Nil.
func (p *TmpfsProvider) Close() error {
	return nil
}
