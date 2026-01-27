// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

// Provider defines the interface for a filesystem provider.
type Provider interface {
	io.Closer

	// GetFs returns the underlying afero.Fs.
	//
	// Returns the result.
	GetFs() afero.Fs

	// ResolvePath resolves a virtual path to the actual path expected by the filesystem.
	//
	// virtualPath is the virtualPath.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ResolvePath(virtualPath string) (string, error)

	// OpenFile opens a file using the provider's filesystem, resolving the path first.
	// It is equivalent to fs.OpenFile(ResolvePath(path), flag, perm) but may include
	// additional security checks (e.g. TOCTOU prevention).
	//
	// path is the virtual path.
	// flag is the open flag.
	// perm is the permission.
	//
	// Returns the file.
	// Returns an error if the operation fails.
	OpenFile(path string, flag int, perm os.FileMode) (afero.File, error)
}
