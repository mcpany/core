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

			s := &searcher{
				fs:              fs,
				ctx:             ctx,
				re:              re,
				excludePatterns: excludePatterns,
				path:            path,
				resolvedPath:    resolvedPath,
				matches:         []map[string]interface{}{},
				maxMatches:      100,
			}

			err = afero.Walk(fs, resolvedPath, s.walkFn)

			if err != nil && err != filepath.SkipDir {
				return nil, err
			}

			return map[string]interface{}{"matches": s.matches}, nil
		},
	}
}

type searcher struct {
	fs              afero.Fs
	ctx             context.Context
	re              *regexp.Regexp
	excludePatterns []string
	path            string
	resolvedPath    string
	matches         []map[string]interface{}
	maxMatches      int
	matchCount      int
}

func (s *searcher) walkFn(filePath string, info os.FileInfo, err error) error {
	if err != nil {
		// Skip unreadable files
		return nil //nolint:nilerr
	}

	// Check context cancellation
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}

	if s.matchCount >= s.maxMatches {
		return filepath.SkipDir
	}

	if s.shouldSkip(info) {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	if info.IsDir() {
		return nil
	}

	return s.processFile(filePath)
}

func (s *searcher) shouldSkip(info os.FileInfo) bool {
	// Check exclusions
	for _, pattern := range s.excludePatterns {
		matched, _ := filepath.Match(pattern, info.Name())
		if matched {
			return true
		}
	}

	// Skip hidden directories like .git and hidden files
	if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".." {
		return true
	}

	// Skip large files (e.g., > 10MB)
	if !info.IsDir() && info.Size() > 10*1024*1024 {
		return true
	}

	return false
}

func (s *searcher) processFile(filePath string) error {
	// Read file
	f, err := s.fs.Open(filePath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	// Check for binary
	if isBinary(f) {
		return nil
	}

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		// Check context inside scanning loop for large files
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}
		lineNum++
		line := scanner.Text()
		if s.re.MatchString(line) {
			s.addMatch(filePath, lineNum, line)
			if s.matchCount >= s.maxMatches {
				return filepath.SkipDir
			}
		}
	}
	return nil
}

func isBinary(f afero.File) bool {
	// Read first 512 bytes
	buffer := make([]byte, 512)
	n, _ := f.Read(buffer)
	if n > 0 {
		contentType := http.DetectContentType(buffer[:n])
		if contentType == "application/octet-stream" {
			return true
		}
		// Reset seeker
		if _, err := f.Seek(0, 0); err != nil {
			return true // Treat seek error as binary/unreadable to skip
		}
	}
	return false
}

func (s *searcher) addMatch(filePath string, lineNum int, line string) {
	// Relativize path
	relPath, _ := filepath.Rel(s.resolvedPath, filePath)
	if relPath == "" {
		relPath = filepath.Base(filePath)
	}

	// Combine with the user-provided path so the result is relative to the provider root
	fullPath := filepath.Join(s.path, relPath)

	s.matches = append(s.matches, map[string]interface{}{
		"file":         fullPath,
		"line_number":  lineNum,
		"line_content": strings.TrimSpace(line),
	})
	s.matchCount++
}
