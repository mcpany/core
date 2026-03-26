// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

// ZipProvider provides access to files within a read-only ZIP archive.
//
// Summary: Implements the FilesystemProvider for ZIP archives.
type ZipProvider struct {
	fs     afero.Fs
	closer *os.File
}

// NewZipProvider creates a new ZipProvider and opens the specified ZIP file.
//
// Summary: Initializes a new ZIP filesystem provider.
//
// Parameters:
//   - config (*configv1.ZipFs): The ZIP configuration containing the file path.
//
// Returns:
//   - *ZipProvider: A pointer to the newly created ZIP provider instance.
//   - error: An error if the ZIP file cannot be opened, validated, or parsed.
//
// Errors:
//   - Returns an error if the ZIP file path is not allowed by egress policy.
//   - Returns an error if the file cannot be opened from the host.
//   - Returns an error if the ZIP reader fails to initialize.
//
// Side Effects:
//   - Opens a file handle to the ZIP archive on the host filesystem.
func NewZipProvider(config *configv1.ZipFs) (*ZipProvider, error) {
	if err := validation.IsAllowedPath(config.GetFilePath()); err != nil {
		return nil, fmt.Errorf("zip file path not allowed: %w", err)
	}

	f, err := os.Open(config.GetFilePath())
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to stat zip file: %w", err)
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	fs := zipfs.New(zr)

	return &ZipProvider{
		fs:     fs,
		closer: f,
	}, nil
}

// GetFs returns the underlying filesystem.
//
// Summary: Retrieves the underlying afero ZIP filesystem.
//
// Returns:
//   - afero.Fs: An afero.Fs implementation backed by the ZIP archive.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Parameters:
//   - None.
func (p *ZipProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a cleaned internal path.
//
// Summary: Cleans and returns the internal path within the ZIP archive.
//
// Parameters:
//   - virtualPath (string): The virtual path relative to the ZIP root.
//
// Returns:
//   - string: The cleaned internal path.
//   - error: Nil, as ZIP paths are internal and relative.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (p *ZipProvider) ResolvePath(virtualPath string) (string, error) {
	// For ZipFs, just clean the path. It's virtual (based on zip contents).
	return filepath.Clean(virtualPath), nil
}

// Close closes the underlying zip file handle.
//
// Summary: Releases the ZIP file resource.
//
// Returns:
//   - error: An error if the file handle closure fails.
//
// Errors:
//   - Returns an error if the underlying os.File close operation fails.
//
// Side Effects:
//   - Closes the file handle on the host filesystem.
//
// Parameters:
//   - None.
func (p *ZipProvider) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}
