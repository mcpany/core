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
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

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

			return executeSearch(ctx, fs, resolvedPath, path, re, excludePatterns)
		},
	}
}

func executeSearch(ctx context.Context, fs afero.Fs, resolvedPath, originalPath string, re *regexp.Regexp, excludePatterns []string) (map[string]interface{}, error) {
	const maxMatches = int32(100)
	var matchCount int32

	// Thread-safe matches collection
	var matchesMu sync.Mutex
	matches := []map[string]interface{}{}

	// Worker pool setup
	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4
	}

	type job struct {
		filePath string
	}

	jobs := make(chan job, numWorkers*2)
	var wg sync.WaitGroup

	// ⚡ BOLT: Concurrent file search using worker pool
	// Randomized Selection from Top 5 High-Impact Targets
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					return
				}
				if atomic.LoadInt32(&matchCount) >= maxMatches {
					return
				}

				processFile(ctx, fs, j.filePath, resolvedPath, originalPath, re, &matchCount, maxMatches, &matches, &matchesMu)
			}
		}()
	}

	// Walker
	walkErr := afero.Walk(fs, resolvedPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip unreadable files
			return nil //nolint:nilerr // Ignore errors accessing individual files
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
		if atomic.LoadInt32(&matchCount) >= maxMatches {
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

		// Send to worker
		select {
		case jobs <- job{filePath: filePath}:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})

	close(jobs)
	wg.Wait()

	if walkErr != nil && walkErr != filepath.SkipDir {
		return nil, walkErr
	}

	return map[string]interface{}{"matches": matches}, nil
}

func processFile(ctx context.Context, fs afero.Fs, filePath, resolvedPath, originalPath string, re *regexp.Regexp, matchCount *int32, maxMatches int32, matches *[]map[string]interface{}, matchesMu *sync.Mutex) {
	f, err := fs.Open(filePath)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	// Check for binary
	buffer := make([]byte, 512)
	n, _ := f.Read(buffer)
	if n > 0 {
		contentType := http.DetectContentType(buffer[:n])
		if contentType == "application/octet-stream" {
			return
		}
		if _, err := f.Seek(0, 0); err != nil {
			return
		}
	}

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}
		if atomic.LoadInt32(matchCount) >= maxMatches {
			return
		}

		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			relPath, _ := filepath.Rel(resolvedPath, filePath)
			if relPath == "" {
				relPath = filepath.Base(filePath)
			}
			fullPath := filepath.Join(originalPath, relPath)

			matchesMu.Lock()
			if len(*matches) < int(maxMatches) {
				*matches = append(*matches, map[string]interface{}{
					"file":         fullPath,
					"line_number":  lineNum,
					"line_content": strings.TrimSpace(line),
				})
				atomic.AddInt32(matchCount, 1)
			}
			matchesMu.Unlock()

			if atomic.LoadInt32(matchCount) >= maxMatches {
				return
			}
		}
	}
}
