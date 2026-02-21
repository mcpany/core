// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
)

func readFileTool(prov provider.Provider, fs afero.Fs) filesystemToolDef {
	return filesystemToolDef{
		Name:        "read_file",
		Description: "Read the content of a file.",
		Input: map[string]interface{}{
			"path": map[string]interface{}{"type": "string", "description": "The path to the file."},
		},
		Output: map[string]interface{}{
			"content": map[string]interface{}{"type": "string", "description": "The file content."},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path is required")
			}

			resolvedPath, err := prov.ResolvePath(path)
			if err != nil {
				return nil, err
			}

			// Check if it's a directory
			info, err := fs.Stat(resolvedPath)
			if err != nil {
				return nil, err
			}
			if info.IsDir() {
				return nil, fmt.Errorf("path is a directory")
			}
			if !info.Mode().IsRegular() {
				return nil, fmt.Errorf("path is not a regular file")
			}

			// Check file size to prevent memory exhaustion (limit to 10MB)
			const maxFileSize = 10 * 1024 * 1024 // 10MB
			if info.Size() > maxFileSize {
				return nil, fmt.Errorf("file size exceeds limit of %d bytes", maxFileSize)
			}

			// Use LimitReader to enforce size limit even if Stat() reported a small size (e.g. /dev/zero, /proc files)
			f, err := fs.Open(resolvedPath)
			if err != nil {
				return nil, err
			}
			defer func() { _ = f.Close() }()

			// Read up to limit + 1 to detect if file is larger (fail-safe for incorrect Stat size)
			reader := io.LimitedReader{R: f, N: maxFileSize + 1}
			contentBytes, err := io.ReadAll(&reader)
			if err != nil {
				return nil, err
			}

			if int64(len(contentBytes)) > maxFileSize {
				return nil, fmt.Errorf("file size exceeds limit of %d bytes", maxFileSize)
			}

			return map[string]interface{}{"content": string(contentBytes)}, nil
		},
	}
}

func writeFileTool(prov provider.Provider, fs afero.Fs, readOnly bool) filesystemToolDef {
	return filesystemToolDef{
		Name:        "write_file",
		Description: "Write content to a file.",
		Input: map[string]interface{}{
			"path":    map[string]interface{}{"type": "string", "description": "The path to the file."},
			"content": map[string]interface{}{"type": "string", "description": "The content to write."},
		},
		Output: map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			if readOnly {
				return nil, fmt.Errorf("filesystem is read-only")
			}
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path is required")
			}
			content, ok := args["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content is required")
			}

			resolvedPath, err := prov.ResolvePath(path)
			if err != nil {
				// Check if parent directory is allowed if file doesn't exist yet
				// resolvePath usually checks validity of prefix.
				return nil, err
			}

			// Ensure parent directory exists
			parentDir := filepath.Dir(resolvedPath)
			if err := fs.MkdirAll(parentDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent directory: %w", err)
			}

			if err := safeWriteFile(fs, resolvedPath, []byte(content), 0600); err != nil {
				return nil, err
			}
			return map[string]interface{}{"success": true}, nil
		},
	}
}

// safeWriteFile writes content to a file safely, preventing TOCTOU attacks.
// It ensures that:
// 1. The path components have not changed (e.g. replaced by symlinks) since resolution.
// 2. If the file does not exist, it is created exclusively (O_EXCL).
// 3. If the file exists, it checks if it is a symlink before opening.
func safeWriteFile(fs afero.Fs, path string, content []byte, perm os.FileMode) error {
	// 1. Re-verify that the path is still canonical (no new symlinks introduced).
	// This mitigates the directory swap attack (TOCTOU) where a parent directory
	// is replaced by a symlink after the initial security check.
	// Note: EvalSymlinks interacts with the OS filesystem. We should only do this check
	// if we are operating on the OS filesystem.
	if _, isOsFs := fs.(*afero.OsFs); isOsFs {
		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			if os.IsNotExist(err) {
				// File doesn't exist. Check parent directory integrity.
				dir := filepath.Dir(path)
				realDir, err := filepath.EvalSymlinks(dir)
				if err != nil {
					return fmt.Errorf("failed to verify parent directory integrity: %w", err)
				}
				if filepath.Clean(realDir) != filepath.Clean(dir) {
					return fmt.Errorf("access denied: path integrity violation (parent directory changed)")
				}
			} else {
				return fmt.Errorf("failed to verify path integrity: %w", err)
			}
		} else {
			// File exists.
			if filepath.Clean(realPath) != filepath.Clean(path) {
				return fmt.Errorf("access denied: path integrity violation (symlink detected)")
			}
		}
	}

	// 2. Try to create the file exclusively first.
	// This covers the case where we are creating a new file.
	// O_CREATE|O_EXCL fails if the file exists.
	f, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err == nil {
		defer f.Close()
		_, err = f.Write(content)
		return err
	}

	if !os.IsExist(err) {
		return err // Real error
	}

	// 3. File exists. We need to overwrite it safely.
	// Race condition: user swaps file with symlink between Lstat and OpenFile.
	// Mitigation: Open-Verify-Truncate dance.

	var lstatInfo os.FileInfo
	if lstater, ok := fs.(afero.Lstater); ok {
		var supported bool
		lstatInfo, supported, err = lstater.LstatIfPossible(path)
		if err != nil {
			return err
		}
		if supported {
			if lstatInfo.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("access denied: cannot overwrite symlink %s", path)
			}
		} else {
			lstatInfo = nil // Not supported, can't verify
		}
	}

	// Open the file WITHOUT truncation first to get a handle to whatever is there.
	f, err = fs.OpenFile(path, os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	// Verify that the file we opened is the same one we Lstat-ed (if we did).
	if lstatInfo != nil {
		fInfo, err := f.Stat()
		if err != nil {
			return err
		}
		if !os.SameFile(lstatInfo, fInfo) {
			return fmt.Errorf("security violation: file identity changed during open (TOCTOU detected)")
		}
	}

	// Now that we verified identity, it is safe to truncate.
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	_, err = f.Write(content)
	return err
}

func moveFileTool(prov provider.Provider, fs afero.Fs, readOnly bool) filesystemToolDef {
	return filesystemToolDef{
		Name:        "move_file",
		Description: "Move or rename a file or directory.",
		Input: map[string]interface{}{
			"source":      map[string]interface{}{"type": "string", "description": "The source path."},
			"destination": map[string]interface{}{"type": "string", "description": "The destination path."},
		},
		Output: map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			if readOnly {
				return nil, fmt.Errorf("filesystem is read-only")
			}
			source, ok := args["source"].(string)
			if !ok {
				return nil, fmt.Errorf("source is required")
			}
			destination, ok := args["destination"].(string)
			if !ok {
				return nil, fmt.Errorf("destination is required")
			}

			resolvedSource, err := prov.ResolvePath(source)
			if err != nil {
				return nil, err
			}

			resolvedDest, err := prov.ResolvePath(destination)
			if err != nil {
				return nil, err
			}

			// Ensure parent of destination exists
			parentDir := filepath.Dir(resolvedDest)
			if err := fs.MkdirAll(parentDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent directory: %w", err)
			}

			if err := fs.Rename(resolvedSource, resolvedDest); err != nil {
				return nil, err
			}
			return map[string]interface{}{"success": true}, nil
		},
	}
}

func deleteFileTool(prov provider.Provider, fs afero.Fs, readOnly bool) filesystemToolDef {
	return filesystemToolDef{
		Name:        "delete_file",
		Description: "Delete a file or directory.",
		Input: map[string]interface{}{
			"path":      map[string]interface{}{"type": "string", "description": "The path to delete."},
			"recursive": map[string]interface{}{"type": "boolean", "description": "Delete recursively if true."},
		},
		Output: map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			if readOnly {
				return nil, fmt.Errorf("filesystem is read-only")
			}
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path is required")
			}
			recursive, _ := args["recursive"].(bool)

			resolvedPath, err := prov.ResolvePath(path)
			if err != nil {
				return nil, err
			}

			if recursive {
				if err := fs.RemoveAll(resolvedPath); err != nil {
					return nil, err
				}
			} else {
				if err := fs.Remove(resolvedPath); err != nil {
					return nil, err
				}
			}
			return map[string]interface{}{"success": true}, nil
		},
	}
}
