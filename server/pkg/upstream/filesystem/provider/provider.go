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
	//
	// Summary: Gets fs.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - Returns error upon failure.
	//
	// Side Effects:
	//   - Interacts with internal state.
	GetFs() afero.Fs

	// ResolvePath resolves a virtual path to the actual path expected by the filesystem.
	//
	// virtualPath is the virtualPath.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	//
	// Summary: Resolves path.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - None.
	//
	// Errors:
	//   - Returns error upon failure.
	//
	// Side Effects:
	//   - Interacts with internal state.
	ResolvePath(virtualPath string) (string, error)
}
