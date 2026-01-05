// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// TmpfsProvider implements the FileProvider interface for temporary memory filesystem.
type TmpfsProvider struct {
	fs afero.Fs
}

// NewTmpfsProvider creates a new TmpfsProvider.
func NewTmpfsProvider() *TmpfsProvider {
	return &TmpfsProvider{
		fs: afero.NewMemMapFs(),
	}
}

// GetFs returns the afero filesystem.
func (p *TmpfsProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to the actual path.
func (p *TmpfsProvider) ResolvePath(virtualPath string) (string, error) {
	// For MemMapFs, just clean the path. It's virtual.
	return filepath.Clean(virtualPath), nil
}

// Close closes the provider.
func (p *TmpfsProvider) Close() error {
	return nil
}
