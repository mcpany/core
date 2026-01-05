// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider is a filesystem provider that uses an in-memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the underlying afero.Fs.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves a virtual path (just cleans it for memfs).
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider (no-op for tmpfs).
func (p *TmpfsProvider) Close() error {
	return nil
}
