// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider provides access to a volatile, temporary in-memory filesystem.
//
// Summary: Implements the FilesystemProvider for in-memory storage (MemMapFs).
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider with an empty memory-backed filesystem.
//
// Summary: Initializes a new in-memory filesystem provider.
//
// Returns:
//   - *TmpfsProvider: A pointer to the newly created TmpfsProvider instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates memory for the in-memory filesystem state.
//
// Parameters:
//   - None.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying filesystem.
//
// Summary: Retrieves the underlying afero in-memory filesystem.
//
// Returns:
//   - afero.Fs: An afero.Fs implementation backed by system memory.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a cleaned internal path.
//
// Summary: Cleans and returns the internal path for the in-memory filesystem.
//
// Parameters:
//   - virtualPath (string): The virtual path provided by the agent.
//
// Returns:
//   - string: The cleaned internal path.
//   - error: Nil, as tmpfs paths are internal and relative.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider and releases any resources.
//
// Summary: Closes the tmpfs provider.
//
// Returns:
//   - error: Nil, as in-memory filesystems don't require explicit closure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
func (p *TmpfsProvider) Close() error {
	return nil
}
