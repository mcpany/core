// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main is a script to remove license headers from files.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// Files to skip.
	skipDirs = map[string]bool{
		"build":        true,
		"vendor":       true,
		"node_modules": true,
		".git":         true,
		".idea":        true,
		".vscode":      true,
	}

	// Extensions to process.
	validExts = map[string]string{
		".go":    "//",
		".py":    "#",
		".sh":    "#",
		".yaml":  "#",
		".yml":   "#",
		".proto": "//",
	}
)

func main() {
	filesProcessed := 0
	filesChanged := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check extension
		ext := filepath.Ext(path)
		prefix, ok := validExts[ext]
		if !ok {
			return nil
		}

		// Skip generated files
		if strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		filesProcessed++
		changed, err := processFile(path, prefix)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
		} else if changed {
			filesChanged++
			fmt.Printf("Removed header from %s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking tree: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processed %d files, removed headers from %d files.\n", filesProcessed, filesChanged)
}

func processFile(path string, prefix string) (bool, error) {
	// G304: Potential file inclusion via variable
	// We are walking the directory tree, so the path is efficient safe, but we clean it to satisfy linter.
	cleanPath := filepath.Clean(path)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return false, nil
	}

	newLines, headerChanged := reconstructContent(lines, prefix)

	if !headerChanged {
		return false, nil
	}

	contentStr := strings.Join(newLines, "\n")
	// Preserve trailing newline if original had it
	if len(lines) > 0 && lines[len(lines)-1] == "" && !strings.HasSuffix(contentStr, "\n") {
		contentStr += "\n"
	}

	// G306: Expect WriteFile permissions to be 0600 or less
	return true, os.WriteFile(cleanPath, []byte(contentStr), 0600)
}

func reconstructContent(lines []string, prefix string) ([]string, bool) {
	var newLines []string
	startIdx := 0

	// 1. Keep shebang
	if strings.HasPrefix(lines[0], "#!") {
		newLines = append(newLines, lines[0])
		startIdx++
	}

	blocks, restStart := identifyBlocks(lines, startIdx, prefix)
	headerChanged := false
	keptBlockCount := 0

	for _, b := range blocks {
		if b.isLicense {
			// If it has build tags, we MUST keep the build tags
			if b.hasBuildTag {
				keptLines := filterBuildTags(b.lines)
				if len(keptLines) > 0 {
					newLines = appendSeparator(newLines)
					newLines = append(newLines, keptLines...)
					keptBlockCount++
				}
				if len(keptLines) < len(b.lines) {
					headerChanged = true
				}
			} else {
				headerChanged = true
			}
		} else {
			newLines = appendSeparator(newLines)
			newLines = append(newLines, b.lines...)
			keptBlockCount++
		}
	}

	// Append rest of file
	restStart = skipLeadingBlankLines(lines, restStart)
	if restStart < len(lines) {
		if len(newLines) > 0 {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, lines[restStart:]...)
	}

	return newLines, headerChanged
}

func identifyBlocks(lines []string, startIdx int, prefix string) ([]Block, int) {
	var blocks []Block
	var currentBlock []string

	i := startIdx
	for ; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(line, prefix)

		if !isComment && trimmed != "" {
			break
		}

		if trimmed == "" {
			if len(currentBlock) > 0 {
				blocks = append(blocks, analyzeBlock(currentBlock))
				currentBlock = nil
			}
			continue
		}
		currentBlock = append(currentBlock, line)
	}

	if len(currentBlock) > 0 {
		blocks = append(blocks, analyzeBlock(currentBlock))
	}
	return blocks, i
}

func skipLeadingBlankLines(lines []string, start int) int {
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	return start
}

func appendSeparator(lines []string) []string {
	if len(lines) > 0 && !strings.HasPrefix(lines[len(lines)-1], "#!") {
		return append(lines, "")
	} else if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "#!") {
		return append(lines, "")
	}
	return lines
}

func filterBuildTags(lines []string) []string {
	var kept []string
	for _, l := range lines {
		if isBuildTag(l) {
			kept = append(kept, l)
		}
	}
	return kept
}

// Block represents a block of comments.
type Block struct {
	lines       []string
	isLicense   bool
	hasBuildTag bool
}

func analyzeBlock(lines []string) (b Block) {
	b.lines = lines
	for _, l := range lines {
		if isBuildTag(l) {
			b.hasBuildTag = true
		}
		lower := strings.ToLower(l)
		if strings.Contains(lower, "copyright") || strings.Contains(lower, "license") || strings.Contains(lower, "apache") {
			// Avoid false positives like "See License for specific language" appearing in package docs without being a header?
			// But usually package docs don't say "Copyright 2025".
			// "Apache License" is strong signal.
			b.isLicense = true
		}
	}
	// If it has build tag but NO license keywords, it's just a build tag block.
	// If matches both, it's a "License Block" (that needs mixed handling).
	return b
}

func isBuildTag(line string) bool {
	trimmed := strings.TrimSpace(line)
	// //go:build ...
	// // +build ...
	if strings.Contains(trimmed, "go:build") || strings.Contains(trimmed, "+build") {
		return true
	}
	return false
}
