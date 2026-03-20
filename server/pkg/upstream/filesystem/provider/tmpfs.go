// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package provider implements the provider subsystem.
package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider provides access to a temporary in-memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider initializes and returns a new tmpfs provider instance.
//
// Parameters:
//   - None
//
// Returns:
//   - *TmpfsProvider: The generated or retrieved entity.
//
// Errors:
//   - None.
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
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs retrieves the fs.
//
// Parameters:
//   - None
//
// Returns:
//   - afero.Fs: The generated or retrieved entity.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// GetFs returns the underlying filesystem.
//
// Returns:
//   - afero.Fs: The result.
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

// Close handles close.
//
// Parameters:
//   - None
//
// Returns:
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
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
func (p *TmpfsProvider) Close() error {
	return nil
}
