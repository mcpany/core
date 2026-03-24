// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// Summary: TmpfsProvider provides access to a temporary in-memory filesystem. Represents a TmpfsProvider.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type TmpfsProvider struct {
	fs afero.Fs
}

// Summary: NewTmpfsProvider creates a new TmpfsProvider.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TmpfsProvider: The resulting *TmpfsProvider.
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

// Summary: GetFs returns the underlying filesystem.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs: The resulting afero.Fs.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// Summary: ResolvePath resolves the virtual path to a real path.
//
// Parameters:
//   - virtualPath (string): The virtualPath parameter.
//
// Returns:
//   - string: The resulting string.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Summary: Close closes the provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
