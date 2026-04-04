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
// Summary: Executes the NewTmpfsProvider operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TmpfsProvider: The returned value.
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
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
// GetFs returns the underlying filesystem.
//
// Summary: Executes the GetFs operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs: The returned value.
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
//
// Summary: Executes ResolvePath operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// ResolvePath resolves the virtual path to a real path.
//
// Summary: Executes the ResolvePath operation.
//
// Parameters:
//   - virtualPath (string): The virtualPath parameter.
//
// Returns:
//   - string: The returned value.
//   - error: The returned value.
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
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
// Close closes the provider.
//
// Summary: Executes the Close operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: The returned value.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
