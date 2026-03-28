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
// Returns:
//   - *TmpfsProvider: The result.
//
// Side Effects:
//   - None.
//
// Summary: Initializes TmpfsProvider with specified constraints.
//
// Parameters:
//   - None
//
// Returns:
//   - {: The resulting {.
//
// Errors:
//   - None
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
// Summary: Retrieves TmpfsProvider with specified constraints.
//
// Parameters:
//   - None
//
// Returns:
//   - {: The resulting {.
//
// Errors:
//   - None
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
// Summary: Executes TmpfsProvider with specified constraints.
//
// Parameters:
//   - virtualPath (string): The virtualPath parameter.
//
// Returns:
//   - string: The resulting string.
//   - {: The resulting {.
//
// Errors:
//   - None
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
// Summary: Executes TmpfsProvider with specified constraints.
//
// Parameters:
//   - None
//
// Returns:
//   - {: The resulting {.
//
// Errors:
//   - None
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
