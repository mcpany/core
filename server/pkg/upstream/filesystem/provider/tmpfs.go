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
//
// Fields:
//   - Contains the configuration and state properties required for TmpfsProvider functionality.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider. Returns: - *TmpfsProvider: The result. Side Effects: - None.
//
// Summary: NewTmpfsProvider creates a new TmpfsProvider. Returns: - *TmpfsProvider: The result. Side Effects: - None.
//
// Parameters:
//   - None.
//
// Returns:
//   - (*TmpfsProvider): The resulting TmpfsProvider object containing the requested data.
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

// GetFs returns the underlying filesystem. Returns: - afero.Fs: The result. Side Effects: - None.
//
// Summary: GetFs returns the underlying filesystem. Returns: - afero.Fs: The result. Side Effects: - None.
//
// Parameters:
//   - None.
//
// Returns:
//   - (afero.Fs): The resulting afero.Fs object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider. Returns: - error: An error if the operation fails. Errors: - Returns an error if ... Side Effects: - None.
//
// Summary: Close closes the provider. Returns: - error: An error if the operation fails. Errors: - Returns an error if ... Side Effects: - None.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
