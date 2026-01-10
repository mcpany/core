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
	if len(p.rootPaths) == 0 {
		return "", fmt.Errorf("no root paths defined")
	}

	// 1. Find best matching root
	bestMatchVirtual, bestMatchReal, err := p.findBestMatch(virtualPath)
	if err != nil {
		return "", err
	}

	// 2. Resolve to canonical real path
	realRootCanonical, targetPathCanonical, err := p.resolvePaths(virtualPath, bestMatchVirtual, bestMatchReal)
	if err != nil {
		return "", err
	}

	// 3. Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	// 4. Allowed/Denied Paths Check
	if err := p.checkAccessControl(targetPathCanonical); err != nil {
		return "", err
	}

	return targetPathCanonical, nil
}

func (p *LocalProvider) findBestMatch(virtualPath string) (string, string, error) {
	var bestMatchVirtual string
	var bestMatchReal string

	for vRoot, rRoot := range p.rootPaths {
		// Ensure vRoot has a clean format
		cleanVRoot := vRoot
		if !strings.HasPrefix(cleanVRoot, "/") {
			cleanVRoot = "/" + cleanVRoot
		}

		// Ensure virtualPath starts with /
		checkPath := virtualPath
		if !strings.HasPrefix(checkPath, "/") {
			checkPath = "/" + checkPath
		}

		if strings.HasPrefix(checkPath, cleanVRoot) {
			// Ensure it matches as a directory component
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
		// Try fallback
		if val, ok := p.rootPaths["/"]; ok {
			bestMatchVirtual = "/"
			bestMatchReal = val
		} else {
			return "", "", fmt.Errorf("path %s is not allowed (no matching root)", virtualPath)
		}
	}
	return bestMatchVirtual, bestMatchReal, nil
}

func (p *LocalProvider) resolvePaths(virtualPath, bestMatchVirtual, bestMatchReal string) (string, string, error) {
	relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
	relativePath = strings.TrimPrefix(relativePath, "/")

	realRootAbs, err := filepath.Abs(bestMatchReal)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve root path: %w", err)
	}

	realRootCanonical, err := filepath.EvalSymlinks(realRootAbs)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve root path symlinks: %w", err)
	}

	targetPath := filepath.Join(realRootCanonical, relativePath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	targetPathCanonical, err := filepath.EvalSymlinks(targetPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			canonical, resolveErr := p.resolveNonExistentPath(targetPathAbs)
			if resolveErr != nil {
				return "", "", resolveErr
			}
			return realRootCanonical, canonical, nil
		}
		return "", "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
	}

	return realRootCanonical, targetPathCanonical, nil
}

func (p *LocalProvider) resolveNonExistentPath(targetPathAbs string) (string, error) {
	currentPath := targetPathAbs
	var existingPath string
	var remainingPath string

	for {
		dir := filepath.Dir(currentPath)
		if dir == currentPath {
			return "", fmt.Errorf("failed to resolve path (root not found): %s", targetPathAbs)
		}

		if _, err := os.Stat(dir); err == nil {
			existingPath = dir
			var relErr error
			remainingPath, relErr = filepath.Rel(existingPath, targetPathAbs)
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

	targetPathCanonical := filepath.Join(existingPathCanonical, remainingPath)
	return targetPathCanonical, nil
}

func (p *LocalProvider) checkAccessControl(targetPath string) error {
	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, pattern := range p.allowedPaths {
			if p.matchPath(pattern, targetPath) {
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
			if p.matchPath(pattern, targetPath) {
				return fmt.Errorf("access denied: path is in denied list")
			}
		}
	}
	return nil
}

func (p *LocalProvider) matchPath(pattern, targetPath string) bool {
	matched, err := filepath.Match(pattern, targetPath)
	if err != nil {
		if strings.HasPrefix(targetPath, pattern) {
			return true
		}
	} else if !matched {
		if strings.HasPrefix(targetPath, pattern) {
			return true
		}
	}
	return matched
}

// Close closes the provider.
func (p *LocalProvider) Close() error {
	return nil
}
