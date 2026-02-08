// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"io"

	"github.com/spf13/afero"
)

// Provider defines the interface for a filesystem provider.
//
// Summary: defines the interface for a filesystem provider.
type Provider interface {
	io.Closer

	// GetFs returns the underlying afero.Fs.
	//
	// Summary: returns the underlying afero.Fs.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - afero.Fs: The afero.Fs.
	GetFs() afero.Fs

	// ResolvePath resolves a virtual path to the actual path expected by the filesystem.
	//
	// Summary: resolves a virtual path to the actual path expected by the filesystem.
	//
	// Parameters:
	//   - virtualPath: string. The string.
	//
	// Returns:
	//   - string: The string.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	ResolvePath(virtualPath string) (string, error)
}
