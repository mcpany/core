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
// Summary: TmpfsProvider provides access to a temporary in-memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
// NewTmpfsProvider creates a new TmpfsProvider.
//
// Summary: NewTmpfsProvider creates a new TmpfsProvider.
//
// Parameters:
//   - None.
//
// Returns:
//   - *TmpfsProvider: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Returns:
//   - *TmpfsProvider: The resulting object or data structure.
//
// Errors:
// GetFs returns the underlying filesystem.
//
// Summary: GetFs returns the underlying filesystem.
//
// Parameters:
//   - None.
//
// Returns:
//   - afero.Fs: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	}
}

// GetFs returns the underlying filesystem.
// ResolvePath resolves the virtual path to a real path.
//
// Summary: ResolvePath resolves the virtual path to a real path.
//
// Parameters:
//   - virtualPath (string): The textual representation of virtualpath.
//
// Returns:
//   - string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path.
// Close closes the provider.
//
// Summary: Close closes the provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
//
// Summary: Close closes the provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (p *TmpfsProvider) Close() error {
	return nil
}
