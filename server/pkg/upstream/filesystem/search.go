// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mcpany/core/server/pkg/upstream/filesystem/provider"
	"github.com/spf13/afero"
)

func searchFilesTool(prov provider.Provider, fs afero.Fs) filesystemToolDef {
	return filesystemToolDef{
		Name:        "search_files",
		Description: "Search for a text pattern in files within a directory.",
		Input: map[string]interface{}{
			"path":             map[string]interface{}{"type": "string", "description": "The root directory to search."},
			"pattern":          map[string]interface{}{"type": "string", "description": "The regular expression to search for."},
			"exclude_patterns": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Glob patterns to exclude (e.g. *.test.js)."},
		},
		Output: map[string]interface{}{
			"matches": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"file":         map[string]interface{}{"type": "string"},
						"line_number":  map[string]interface{}{"type": "integer"},
						"line_content": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path is required")
			}
			patternStr, ok := args["pattern"].(string)
			if !ok {
				return nil, fmt.Errorf("pattern is required")
			}

			re, err := regexp.Compile(patternStr)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern: %w", err)
			}

			var excludePatterns []string
			if ep, ok := args["exclude_patterns"].([]interface{}); ok {
				for _, p := range ep {
					if ps, ok := p.(string); ok {
						excludePatterns = append(excludePatterns, ps)
					}
				}
			}

			resolvedPath, err := prov.ResolvePath(path)
			if err != nil {
				return nil, err
			}

			matches := []map[string]interface{}{}
			maxMatches := 100
			matchCount := 0

			err = afero.Walk(fs, resolvedPath, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					// Skip unreadable files
					return nil
				}

				// Check context cancellation
				if ctx.Err() != nil {
					return ctx.Err()
				}

				if matchCount >= maxMatches {
					return filepath.SkipDir
				}

				// Check exclusions
				for _, pattern := range excludePatterns {
					matched, _ := filepath.Match(pattern, info.Name())
					if matched {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
				}

				if info.IsDir() {
					// Skip hidden directories like .git
					if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".." {
						return filepath.SkipDir
					}
					return nil
				}
				// Skip large files (e.g., > 10MB)
				if info.Size() > 10*1024*1024 {
					return nil
				}

				// Read file
				f, err := fs.Open(filePath)
				if err != nil {
					return nil
				}
				defer func() { _ = f.Close() }()

				// Check for binary
				// Read first 512 bytes
				buffer := make([]byte, 512)
				n, _ := f.Read(buffer)
				if n > 0 {
					contentType := http.DetectContentType(buffer[:n])
					if contentType == "application/octet-stream" {
						return nil
					}
					// Reset seeker
					if _, err := f.Seek(0, 0); err != nil {
						return nil
					}
				}

				scanner := bufio.NewScanner(f)
				lineNum := 0
				for scanner.Scan() {
					// Check context inside scanning loop for large files
					if ctx.Err() != nil {
						return ctx.Err()
					}
					lineNum++
					line := scanner.Text()
					if re.MatchString(line) {
						// Relativize path
						relPath, _ := filepath.Rel(resolvedPath, filePath)
						if relPath == "" {
							relPath = filepath.Base(filePath)
						}

						// Combine with the user-provided path so the result is relative to the provider root (or absolute if input was absolute)
						// resolvedPath corresponds to path.
						fullPath := filepath.Join(path, relPath)

						matches = append(matches, map[string]interface{}{
							"file":         fullPath,
							"line_number":  lineNum,
							"line_content": strings.TrimSpace(line),
						})
						matchCount++
						if matchCount >= maxMatches {
							return filepath.SkipDir
						}
					}
				}
				return nil
			})

			if err != nil && err != filepath.SkipDir {
				return nil, err
			}

			return map[string]interface{}{"matches": matches}, nil
		},
	}
}
