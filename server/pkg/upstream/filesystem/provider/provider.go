// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"io"

	"github.com/spf13/afero"
)

// Provider defines the interface for a filesystem provider.
//
// Summary: Represents a Provider.
type Provider interface {
	io.Closer

	// GetFs returns the underlying afero.Fs.
	//
	// Returns the result.
	// GetFs ...
	//
	// Summary: Retrieves GetFs operation.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - afero.Fs: The afero.Fs result.
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	GetFs() afero.Fs

	// ResolvePath resolves a virtual path to the actual path expected by the filesystem.
	//
	// virtualPath is the virtualPath.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	// ResolvePath ...
	//
	// Summary: Executes ResolvePath operation.
	//
	// Parameters:
	//   - virtualPath: string. A string value.
	//
	// Returns:
	//   - string: The resulting string.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None.
	ResolvePath(virtualPath string) (string, error)
}
