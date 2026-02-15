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
	symlinkMode  configv1.FilesystemUpstreamService_SymlinkMode
}

// NewLocalProvider creates a new LocalProvider from the given configuration.
//
// _ is an unused parameter.
// rootPaths is the rootPaths.
// allowedPaths is the allowedPaths.
// deniedPaths is the deniedPaths.
// symlinkMode is the symlinkMode.
//
// Returns the result.
func NewLocalProvider(_ *configv1.OsFs, rootPaths map[string]string, allowedPaths, deniedPaths []string, symlinkMode configv1.FilesystemUpstreamService_SymlinkMode) *LocalProvider {
	return &LocalProvider{
		fs:           afero.NewOsFs(),
		rootPaths:    rootPaths,
		allowedPaths: allowedPaths,
		deniedPaths:  deniedPaths,
		symlinkMode:  symlinkMode,
	}
}

// GetFs returns the underlying filesystem.
//
// Returns the result.
func (p *LocalProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the local filesystem.
//
// virtualPath is the virtualPath.
//
// Returns the result.
// Returns an error if the operation fails.
func (p *LocalProvider) ResolvePath(virtualPath string) (string, error) {
	if len(p.rootPaths) == 0 {
		return "", fmt.Errorf("no root paths defined")
	}

	bestMatchVirtual, bestMatchReal, err := p.findBestMatch(virtualPath)
	if err != nil {
		return "", err
	}

	// For DENY mode, check if the input path relative to root contains any symlinks
	if p.symlinkMode == configv1.FilesystemUpstreamService_DENY {
		if contains, err := p.containsSymlink(virtualPath, bestMatchVirtual, bestMatchReal); err != nil {
			return "", err
		} else if contains {
			return "", fmt.Errorf("access denied: symlinks are disabled")
		}
	}

	targetPathCanonical, realRootCanonical, err := p.resolveSymlinks(virtualPath, bestMatchVirtual, bestMatchReal)
	if err != nil {
		return "", err
	}

	if err := p.checkPathSecurity(targetPathCanonical, realRootCanonical); err != nil {
		return "", err
	}

	return targetPathCanonical, nil
}

func (p *LocalProvider) findBestMatch(virtualPath string) (string, string, error) {
	var bestMatchVirtual string
	var bestMatchReal string

	for vRoot, rRoot := range p.rootPaths {
		// Ensure vRoot has a clean format
		cleanVRoot := strings.TrimSuffix(vRoot, "/")
		if !strings.HasPrefix(cleanVRoot, "/") {
			cleanVRoot = "/" + cleanVRoot
		}

		// Ensure virtualPath starts with /
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

func (p *LocalProvider) resolveSymlinks(virtualPath, bestMatchVirtual, bestMatchReal string) (string, string, error) {
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
			targetPathCanonical, err = p.resolveNonExistentPath(targetPathAbs)
			if err != nil {
				return "", "", err
			}
		} else {
			return "", "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
		}
	}
	return targetPathCanonical, realRootCanonical, nil
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

	// Safety check: ensure we are not traversing any broken symlinks in the remaining path.
	// existingPath is the deepest ancestor that Stat() confirmed exists.
	// Therefore, any existing component in remainingPath must be either:
	// 1. A broken symlink (Stat failed, but Lstat might succeed)
	// 2. A file/dir with permission denied (Stat failed)
	// We want to block (1).
	if remainingPath != "." && remainingPath != "" {
		parts := strings.Split(remainingPath, string(os.PathSeparator))
		if len(parts) > 0 {
			nextComponent := parts[0]
			checkPath := filepath.Join(existingPath, nextComponent)
			info, err := os.Lstat(checkPath)
			if err == nil {
				// The component exists (at least to Lstat).
				// Since Stat failed (or we wouldn't be in this part of the loop),
				// checking if it is a symlink is sufficient to identify a broken symlink.
				if info.Mode()&os.ModeSymlink != 0 {
					return "", fmt.Errorf("access denied: component %s is a broken symlink", nextComponent)
				}
			}
		}
	}

	return filepath.Join(existingPathCanonical, remainingPath), nil
}

func (p *LocalProvider) checkPathSecurity(targetPathCanonical, realRootCanonical string) error {
	// Robust check using filepath.Rel to prevent prefix matching issues (e.g. /vol/foo vs /vol/foobar)
	rel, err := filepath.Rel(realRootCanonical, targetPathCanonical)
	if err != nil {
		return fmt.Errorf("access denied: failed to calculate relative path: %w", err)
	}

	// If relative path starts with ".." or is absolute, it's outside.
	// Also check if we are on different volume (Windows)
	if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
		return fmt.Errorf("access denied: path traversal detected")
	}

	// Double check strict prefix as defense in depth (handling trailing slash behavior)
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return fmt.Errorf("access denied: path traversal detected (prefix check)")
	}

	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, pattern := range p.allowedPaths {
			if p.matchPath(pattern, targetPathCanonical) {
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
			if p.matchPath(pattern, targetPathCanonical) {
				return fmt.Errorf("access denied: path is in denied list")
			}
		}
	}

	return nil
}

func (p *LocalProvider) matchPath(pattern, targetPath string) bool {
	matched, err := filepath.Match(pattern, targetPath)
	if err == nil && matched {
		return true
	}
	return strings.HasPrefix(targetPath, pattern)
}

func (p *LocalProvider) containsSymlink(virtualPath, bestMatchVirtual, bestMatchReal string) (bool, error) {
	relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
	relativePath = strings.TrimPrefix(relativePath, "/")

	if relativePath == "" || relativePath == "." {
		return false, nil
	}

	current := bestMatchReal
	parts := strings.Split(relativePath, string(os.PathSeparator))

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}

	return false, nil
}

// Close closes the provider.
//
// Returns an error if the operation fails.
func (p *LocalProvider) Close() error {
	return nil
}
