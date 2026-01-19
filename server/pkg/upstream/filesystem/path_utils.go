// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

func (u *Upstream) validateLocalPaths(rootPaths map[string]string) {
	log := logging.GetLogger()
	for virtualPath, localPath := range rootPaths {
		normalizedPath := normalizePath(localPath)

		// Check if normalization changed the path
		if normalizedPath != localPath {
			log.Info("Filesystem upstream: normalized root path", "original", localPath, "normalized", normalizedPath)
			rootPaths[virtualPath] = normalizedPath
		}

		// Use os.Stat to check if the path exists
		_, err := os.Stat(normalizedPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn("Filesystem upstream: configured root path does not exist, removing from configuration", "virtual_path", virtualPath, "local_path", localPath, "normalized_path", normalizedPath)
			} else {
				log.Warn("Filesystem upstream: error accessing configured root path, removing from configuration", "virtual_path", virtualPath, "local_path", localPath, "error", err)
			}
			delete(rootPaths, virtualPath)
		}
	}
}

func normalizePath(path string) string {
	// 1. Handle file:// URI scheme
	if strings.HasPrefix(path, "file://") {
		// Parse the URL
		u, err := url.Parse(path)
		if err == nil {
			path = u.Path
			// For Windows, /C:/path needs to be C:/path (or C:\path)
			// url.Parse might return /C:/path on Windows for file:///C:/path
			if runtime.GOOS == "windows" {
				// Remove leading slash if it precedes a drive letter
				if len(path) > 2 && path[0] == '/' && path[2] == ':' {
					path = path[1:]
				}
				// Also URL decoding might be needed if not handled by Parse?
				// url.Parse decodes %20 but maybe not everything if we just take u.Path?
				// Actually u.Path is already decoded.
			}
		} else {
            // Fallback: manual stripping if parse fails (unlikely for simple file://)
			path = strings.TrimPrefix(path, "file://")
		}
	}

	// 2. Handle URL encoding (e.g. %20, %3A) manually if not covered by url.Parse for raw strings
    // If the string was NOT a URL, we might still have %3A or similar if user copy-pasted wrong.
    // But be careful not to decode valid names containing %.
    // Generally, if it wasn't a file:// URI, we assume it's a file system path.
    // However, the issue report mentioned `file:///c%3A/temp`. url.Parse handles this.

	// 3. Normalize separators for the OS
	path = filepath.Clean(path)

	return path
}
