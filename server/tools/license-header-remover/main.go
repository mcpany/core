// Package main implements a tool to remove license headers.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Split the regex pattern to avoid matching this source file.
var (
	copyrightPattern = `(?i)copyright`
	copyrightRegex   = regexp.MustCompile(copyrightPattern)
	shebangRegex     = regexp.MustCompile(`^#!`)
	spdxRegex        = regexp.MustCompile(`SPDX-License-Identifier`)
	limitationsRegex = regexp.MustCompile(`limitations under the License`)
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			// Skip hidden directories and build/vendor dirs
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "build" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip generated protobuf files
		if strings.HasSuffix(path, ".pb.go") {
			return nil
		}

		// Process source files
		if isSourceFile(path) {
			processFile(path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

func isSourceFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".py", ".sh", ".yaml", ".yml", ".proto":
		return true
	}
	base := filepath.Base(path)
	if base == "Makefile" || base == "Dockerfile" {
		return true
	}
	return false
}

func processFile(path string) {
	content, err := os.ReadFile(path) //nolint:gosec // CLI tool
	if err != nil {
		fmt.Printf("Failed to read %s: %v\n", path, err)
		return
	}

	lines := strings.Split(string(content), "\n")

	// Find the line with copyright
	copyrightLineIdx := -1
	for i, line := range lines {
		if copyrightRegex.MatchString(line) {
			// Check if this looks like a comment
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
				copyrightLineIdx = i
				break
			}
		}
	}

	if copyrightLineIdx == -1 {
		return // No license header found
	}

	// Detect comment style based on the copyright line
	line := lines[copyrightLineIdx]
	trimmed := strings.TrimSpace(line)

	var startIdx, endIdx int

	switch {
	case strings.HasPrefix(trimmed, "//"):
		startIdx, endIdx = findBlock(lines, copyrightLineIdx, "//")
		endIdx = refineEndIndex(lines, startIdx, endIdx)
	case strings.HasPrefix(trimmed, "#"):
		startIdx, endIdx = findBlock(lines, copyrightLineIdx, "#")
		endIdx = refineEndIndex(lines, startIdx, endIdx)
	case strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*"):
		// Assume C-style block comment
		startIdx, endIdx = findBlockComment(lines, copyrightLineIdx)
		// For /* */ we do not refine end index based on content, because we must remove until */
	default:
		// Should not happen due to check above
		return
	}

	// Validate block
	if startIdx == -1 || endIdx == -1 {
		return
	}

	// Safety check: Ensure the block is at the top of the file
	if !isHeaderBlock(lines, startIdx) {
		fmt.Printf("Skipping %s: Copyright found at line %d but not at file header\n", path, copyrightLineIdx+1)
		return
	}

	fmt.Printf("Removing license header from %s (lines %d-%d)\n", path, startIdx+1, endIdx+1)

	// Remove lines
	var newLines []string
	newLines = append(newLines, lines[:startIdx]...)
	if endIdx+1 < len(lines) {
		newLines = append(newLines, lines[endIdx+1:]...)
	}

	// Write back
	output := strings.Join(newLines, "\n")
	info, _ := os.Stat(path)
	err = os.WriteFile(path, []byte(output), info.Mode())
	if err != nil {
		fmt.Printf("Failed to write %s: %v\n", path, err)
	}
}

func findBlock(lines []string, idx int, prefix string) (int, int) {
	start := idx
	end := idx

	// Scan up
	for i := idx; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, prefix) {
			break
		}
		if prefix == "#" && shebangRegex.MatchString(lines[i]) {
			break
		}
		start = i
	}

	// Scan down
	for i := idx; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, prefix) {
			break
		}
		end = i
	}

	return start, end
}

func refineEndIndex(lines []string, start, end int) int {
	lastMarkerIdx := -1
	for i := start; i <= end; i++ {
		if spdxRegex.MatchString(lines[i]) || limitationsRegex.MatchString(lines[i]) {
			lastMarkerIdx = i
		}
	}
	if lastMarkerIdx != -1 {
		return lastMarkerIdx
	}
	return end
}

func findBlockComment(lines []string, idx int) (int, int) {
	start := idx
	end := idx

	// Scan up for "/*"
	foundStart := false
	for i := idx; i >= 0; i-- {
		if strings.Contains(lines[i], "/*") {
			start = i
			foundStart = true
			break
		}
	}
	if !foundStart {
		return -1, -1
	}

	// Scan down for "*/"
	foundEnd := false
	for i := idx; i < len(lines); i++ {
		if strings.Contains(lines[i], "*/") {
			end = i
			foundEnd = true
			break
		}
	}
	if !foundEnd {
		return -1, -1
	}

	return start, end
}

// isHeaderBlock checks if the lines before startIdx are only empty lines, shebangs, or other comments.
// It returns true if it looks like a header block.
func isHeaderBlock(lines []string, startIdx int) bool {
	for i := 0; i < startIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if shebangRegex.MatchString(lines[i]) {
			continue
		}
		// If there are other comments before, that's fine (e.g. build tags)
		// But if there is code, return false.
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
			continue
		}
		// If we encounter "package" or "import" or anything else, return false
		return false
	}
	return true
}
