// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
)

// LocalProvider provides access to the local filesystem.
type LocalProvider struct {
	fs           afero.Fs
	rootPaths    map[string]string
	allowedPaths []string
	deniedPaths  []string
}

// NewLocalProvider creates a new LocalProvider from the given configuration.
func NewLocalProvider(_ *configv1.OsFs, rootPaths map[string]string, allowedPaths, deniedPaths []string) *LocalProvider {
	return &LocalProvider{
		fs:           afero.NewOsFs(),
		rootPaths:    rootPaths,
		allowedPaths: allowedPaths,
		deniedPaths:  deniedPaths,
	}
}

// GetFs returns the underlying filesystem.
func (p *LocalProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the local filesystem.
func (p *LocalProvider) ResolvePath(virtualPath string) (string, error) {
	bestMatchVirtual, bestMatchReal, err := p.resolveRoot(virtualPath)
	if err != nil {
		return "", err
	}

	// 2. Resolve the path
	relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
	relativePath = strings.TrimPrefix(relativePath, "/")

	realRootAbs, err := filepath.Abs(bestMatchReal)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root path: %w", err)
	}

	realRootCanonical, err := filepath.EvalSymlinks(realRootAbs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root path symlinks: %w", err)
	}

	targetPath := filepath.Join(realRootCanonical, relativePath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	targetPathCanonical, err := resolveSymlinks(targetPathAbs)
	if err != nil {
		return "", err
	}

	// 3. Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	// 4. Allowed/Denied Paths Check
	if err := p.checkAccess(targetPathCanonical); err != nil {
		return "", err
	}

	return targetPathCanonical, nil
}

// Close closes the provider.
func (p *LocalProvider) Close() error {
	return nil
}
