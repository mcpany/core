// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: TmpfsProvider provides access to a temporary in-memory filesystem.
//
// Side Effects:
//   - None.
//
// Summary: NewTmpfsProvider creates a new TmpfsProvider.
//
// Returns:
//   - *TmpfsProvider: The result.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Summary: GetFs returns the underlying filesystem.
//
// Returns:
//   - afero.Fs: The result.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Summary: ResolvePath resolves the virtual path to a real path.
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
	// Summary: Close closes the provider.
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
	// Parameters:
	//   - None.
	return filepath.Clean(virtualPath), nil
}

func (p *TmpfsProvider) Close() error {
	return nil
}
