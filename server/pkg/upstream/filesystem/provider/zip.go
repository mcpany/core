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

// ZipProvider provides access to files within a zip archive.
//
// Summary: Represents a ZipProvider.
type ZipProvider struct {
	fs     afero.Fs
	closer *os.File
}

// NewZipProvider creates a new ZipProvider from the given configuration.
//
// Parameters:
//   - config (*configv1.ZipFs): The parameter.
//
// Returns:
//   - *ZipProvider: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if ...
//
// Side Effects:
//   - None.
//
// Summary: Initializes NewZipProvider operation.
//
// Parameters:
//   - config (*configv1.ZipFs): The zip configuration.
//
// Returns:
//   - *ZipProvider: The new zip provider instance.
//   - error: An error if initialization fails.
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
// Returns:
//   - afero.Fs: The result.
//
// Side Effects:
//   - None.
//
// Summary: Retrieves GetFs operation.
//
// Returns:
//   - afero.Fs: The underlying afero filesystem.
func (p *ZipProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the zip.
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
// Summary: Executes ResolvePath operation.
//
// Parameters:
//   - virtualPath (string): The virtual path to resolve.
//
// Returns:
//   - string: The resolved local path.
//   - error: An error if resolution fails.
func (p *ZipProvider) ResolvePath(virtualPath string) (string, error) {
	// For ZipFs, just clean the path. It's virtual (based on zip contents).
	return filepath.Clean(virtualPath), nil
}

// Close closes the underlying zip file.
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
// Summary: Executes Close operation.
//
// Returns:
//   - error: An error if closure fails.
func (p *ZipProvider) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}
