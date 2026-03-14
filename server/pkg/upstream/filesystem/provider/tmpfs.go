// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: TmpfsProvider provides access to a temporary in-memory filesystem.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// NewTmpfsProvider creates a new TmpfsProvider.
//
// Returns:
//   - *TmpfsProvider: The result.
//
// Side Effects:
//   - None.
//
//
// Errors:
//   - An error if it fails.
// GetFs returns the underlying filesystem.
//
// Returns:
//   - afero.Fs: The result.
//
// Side Effects:
//   - None.
//
//
// Errors:
//   - An error if it fails.
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
package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

type TmpfsProvider struct {
	fs afero.Fs
}

func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
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
	return filepath.Clean(virtualPath), nil
}

func (p *TmpfsProvider) Close() error {
	return nil
}
