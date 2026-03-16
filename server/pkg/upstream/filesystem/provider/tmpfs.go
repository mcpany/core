// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider provides access to a temporary in-memory filesystem.
//
// Summary: TmpfsProvider provides access to a temporary in-memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider.
//
// Summary: NewTmpfsProvider creates a new TmpfsProvider.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TmpfsProvider: The *TmpfsProvider result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
// Summary: GetFs returns the underlying filesystem.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs: The afero.Fs result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
// Summary: ResolvePath resolves the virtual path to a real path.
//
// Parameters:
//   - virtualPath (string): The virtualPath parameter.
//
// Returns:
//   - string: The string result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
//
// Summary: Close closes the provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (p *TmpfsProvider) Close() error {
	return nil
}
