// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"context"
	"fmt"
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

			content, err := afero.ReadFile(fs, resolvedPath)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"content": string(content)}, nil
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

			if err := afero.WriteFile(fs, resolvedPath, []byte(content), 0600); err != nil {
				return nil, err
			}
			return map[string]interface{}{"success": true}, nil
		},
	}
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
