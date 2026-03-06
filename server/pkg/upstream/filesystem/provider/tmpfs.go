// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider - Auto-generated documentation.
//
// Summary: TmpfsProvider provides access to a temporary in-memory filesystem.
//
// Fields:
//   - Various fields for TmpfsProvider.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider - Auto-generated documentation.
//
// Summary: NewTmpfsProvider creates a new TmpfsProvider.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs - Auto-generated documentation.
//
// Summary: GetFs returns the underlying filesystem.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// Close - Auto-generated documentation.
//
// Summary: Close closes the provider.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (p *TmpfsProvider) Close() error {
	return nil
}
