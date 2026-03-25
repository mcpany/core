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

// NewZipProvider creates a new zip provider.
//
// Summary: Creates a new zip provider.
//
// Parameters: - None.
//   - config (*configv1.ZipFs): The config.
//
// Returns: - None.
//   - *ZipProvider: The result.
//   - error: An error if the operation fails.
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

// GetFs retrieves the fs.
//
// Summary: Retrieves the fs.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - afero.Fs: The result.
func (p *ZipProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolvePath resolve path.
//
// Summary: ResolvePath resolve path.
//
// Parameters: - None.
//   - virtualPath (string): The virtual path.
//
// Returns: - None.
//   - string: The result.
//   - error: An error if the operation fails.
func (p *ZipProvider) ResolvePath(virtualPath string) (string, error) {
	// For ZipFs, just clean the path. It's virtual (based on zip contents).
	return filepath.Clean(virtualPath), nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (p *ZipProvider) Close() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}
