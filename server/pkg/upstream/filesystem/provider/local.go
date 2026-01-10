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
// ResolvePath resolves the virtual path to a real path in the local filesystem.
func (p *LocalProvider) ResolvePath(virtualPath string) (string, error) {
	if len(p.rootPaths) == 0 {
		return "", fmt.Errorf("no root paths defined")
	}

	bestMatchVirtual, bestMatchReal, err := p.findBestRoot(virtualPath)
	if err != nil {
		return "", err
	}

	// Resolve the path
	relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
	// handle case where virtualPath matched exactly or with trailing slash
	relativePath = strings.TrimPrefix(relativePath, "/")

	realRootAbs, err := filepath.Abs(bestMatchReal)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Resolve symlinks in the root path to ensure we have the canonical path
	realRootCanonical, err := filepath.EvalSymlinks(realRootAbs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root path symlinks: %w", err)
	}

	targetPath := filepath.Join(realRootCanonical, relativePath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Resolve symlinks for the target path too
	targetPathCanonical, err := p.resolveSymlinks(targetPathAbs)
	if err != nil {
		return "", err
	}

	// Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	if err := p.checkPathAccess(targetPathCanonical); err != nil {
		return "", err
	}

	return targetPathCanonical, nil
}

func (p *LocalProvider) findBestRoot(virtualPath string) (string, string, error) {
	var bestMatchVirtual string
	var bestMatchReal string

	for vRoot, rRoot := range p.rootPaths {
		cleanVRoot := vRoot
		if !strings.HasPrefix(cleanVRoot, "/") {
			cleanVRoot = "/" + cleanVRoot
		}

		checkPath := virtualPath
		if !strings.HasPrefix(checkPath, "/") {
			checkPath = "/" + checkPath
		}

		if strings.HasPrefix(checkPath, cleanVRoot) {
			validMatch := checkPath == cleanVRoot ||
				strings.HasPrefix(checkPath, cleanVRoot+"/") ||
				cleanVRoot == "/"

			if validMatch {
				if len(cleanVRoot) > len(bestMatchVirtual) {
					bestMatchVirtual = cleanVRoot
					bestMatchReal = rRoot
				}
			}
		}
	}

	if bestMatchVirtual == "" {
		if val, ok := p.rootPaths["/"]; ok {
			bestMatchVirtual = "/"
			bestMatchReal = val
		} else {
			return "", "", fmt.Errorf("path %s is not allowed (no matching root)", virtualPath)
		}
	}
	return bestMatchVirtual, bestMatchReal, nil
}

func (p *LocalProvider) resolveSymlinks(path string) (string, error) {
	targetPathCanonical, err := filepath.EvalSymlinks(path)
	if err == nil {
		return targetPathCanonical, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
	}

	// If file doesn't exist, we need to find the deepest existing ancestor
	currentPath := path
	var existingPath string
	var remainingPath string

	for {
		dir := filepath.Dir(currentPath)
		if dir == currentPath {
			return "", fmt.Errorf("failed to resolve path (root not found): %s", path)
		}

		if _, err := os.Stat(dir); err == nil {
			existingPath = dir
			var relErr error
			remainingPath, relErr = filepath.Rel(existingPath, path)
			if relErr != nil {
				return "", fmt.Errorf("failed to calculate relative path: %w", relErr)
			}
			break
		}
		currentPath = dir
	}

	existingPathCanonical, err := filepath.EvalSymlinks(existingPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve ancestor path symlinks: %w", err)
	}

	return filepath.Join(existingPathCanonical, remainingPath), nil
}

func (p *LocalProvider) checkPathAccess(path string) error {
	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, pattern := range p.allowedPaths {
			if p.matchPattern(pattern, path) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access denied: path not in allowed list")
		}
	}

	if len(p.deniedPaths) > 0 {
		for _, pattern := range p.deniedPaths {
			if p.matchPattern(pattern, path) {
				return fmt.Errorf("access denied: path is in denied list")
			}
		}
	}
	return nil
}

func (p *LocalProvider) matchPattern(pattern, path string) bool {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return strings.HasPrefix(path, pattern)
	}
	if !matched {
		return strings.HasPrefix(path, pattern)
	}
	return true
}

// Close closes the provider.
func (p *LocalProvider) Close() error {
	return nil
}
