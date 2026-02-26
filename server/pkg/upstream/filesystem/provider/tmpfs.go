// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider provides access to a temporary in-memory filesystem.
//
// Summary: Provides access to a temporary in-memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider.
//
//
// Summary: Creates a new TmpfsProvider.
//
// Returns:
// - *TmpfsProvider: The result.
//
// Side Effects:
// - None.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
//
// Summary: Returns the underlying filesystem.
//
// Returns:
// - afero.Fs: The result.
//
// Side Effects:
// - None.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
//
//
// Summary: Resolves the virtual path to a real path.
//
// Parameters:
// - virtualPath (string): The string value.
//
// Returns:
// - string: The result.
// - error: An error if the operation fails.
//
// Errors:
// - Returns an error if ...
//
// Side Effects:
// - None.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
//
//
// Summary: Closes the provider.
//
// Returns:
// - error: An error if the operation fails.
//
// Errors:
// - Returns an error if ...
//
// Side Effects:
// - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
