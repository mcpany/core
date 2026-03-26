// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider tmpfsProvider represents a tmpfs provider.
//
// Summary: TmpfsProvider represents a tmpfs provider.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider.
//
// Returns: - None.
//   - *TmpfsProvider: The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewTmpfsProvider operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
// Returns: - None.
//   - afero.Fs: The result.
//
// Side Effects: - None.
//   - None.
//
// Summary: Retrieves GetFs operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
// Parameters: - None.
//   - virtualPath (string): The parameter.
//
// Returns: - None.
//   - string: The result.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if ...
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes ResolvePath operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
//
// Returns: - None.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if ...
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Close operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
