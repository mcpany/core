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
//
// Parameters:
//   _ : The OsFs configuration (unused).
//   rootPaths: A map of virtual roots to real roots.
//   allowedPaths: A list of allowed path patterns.
//   deniedPaths: A list of denied path patterns.
//
// Returns:
//   *LocalProvider: A new LocalProvider instance.
func NewLocalProvider(_ *configv1.OsFs, rootPaths map[string]string, allowedPaths, deniedPaths []string) *LocalProvider {
	return &LocalProvider{
		fs:           afero.NewOsFs(),
		rootPaths:    rootPaths,
		allowedPaths: allowedPaths,
		deniedPaths:  deniedPaths,
	}
}

// GetFs returns the underlying filesystem.
//
// Returns:
//   afero.Fs: The filesystem interface.
func (p *LocalProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the local filesystem.
// It handles root mapping, symlink resolution, and access control checks.
//
// Parameters:
//   virtualPath: The virtual path to resolve.
//
// Returns:
//   string: The resolved absolute canonical path.
//   error: An error if resolution or access check fails.
func (p *LocalProvider) ResolvePath(virtualPath string) (string, error) {
	if len(p.rootPaths) == 0 {
		return "", fmt.Errorf("no root paths defined")
	}

	bestMatchVirtual, bestMatchReal, err := p.findBestMatchRoot(virtualPath)
	if err != nil {
		return "", err
	}

	targetPathCanonical, err := p.resolveTargetCanonical(virtualPath, bestMatchVirtual, bestMatchReal)
	if err != nil {
		return "", err
	}

	// Allowed/Denied Paths Check
	if err := p.checkAccess(targetPathCanonical); err != nil {
		return "", err
	}

	return targetPathCanonical, nil
}

func (p *LocalProvider) findBestMatchRoot(virtualPath string) (string, string, error) {
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
		// Try fallback: if rootPaths has "/" key, use it.
		if val, ok := p.rootPaths["/"]; ok {
			bestMatchVirtual = "/"
			bestMatchReal = val
		} else {
			return "", "", fmt.Errorf("path %s is not allowed (no matching root)", virtualPath)
		}
	}
	return bestMatchVirtual, bestMatchReal, nil
}

func (p *LocalProvider) resolveTargetCanonical(virtualPath, bestMatchVirtual, bestMatchReal string) (string, error) {
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

	targetPathCanonical, err := filepath.EvalSymlinks(targetPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return p.resolveNonExistentPath(targetPathAbs, realRootCanonical)
		}
		return "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
	}

	// Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	return targetPathCanonical, nil
}

func (p *LocalProvider) resolveNonExistentPath(targetPathAbs, realRootCanonical string) (string, error) {
	currentPath := targetPathAbs
	var existingPath string
	var remainingPath string

	for {
		dir := filepath.Dir(currentPath)
		if dir == currentPath {
			// Reached root without finding existing path (should not happen if realRoot exists)
			return "", fmt.Errorf("failed to resolve path (root not found): %s", targetPathAbs)
		}

		// Check if dir exists
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

	// Resolve symlinks for the existing ancestor
	existingPathCanonical, err := filepath.EvalSymlinks(existingPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve ancestor path symlinks: %w", err)
	}

	// Construct the canonical path
	targetPathCanonical := filepath.Join(existingPathCanonical, remainingPath)

	// Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	return targetPathCanonical, nil
}

func (p *LocalProvider) checkAccess(targetPathCanonical string) error {
	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, pattern := range p.allowedPaths {
			matched, err := filepath.Match(pattern, targetPathCanonical)
			if err != nil {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			} else if !matched {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			}

			if matched {
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
			matched, err := filepath.Match(pattern, targetPathCanonical)
			if err != nil {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			} else if !matched {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			}

			if matched {
				return fmt.Errorf("access denied: path is in denied list")
			}
		}
	}
	return nil
}

// Close closes the provider.
//
// Returns:
//   error: nil.
func (p *LocalProvider) Close() error {
	return nil
}
