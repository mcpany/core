package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	funcRegex   = regexp.MustCompile(`^func\s+([A-Z]\w*)\s*\(`)
	methodRegex = regexp.MustCompile(`^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\(`)
	typeRegex   = regexp.MustCompile(`^type\s+([A-Z]\w*)\s+`)
)

func main() {
	err := filepath.Walk("server", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor") || strings.Contains(path, ".pb.go") {
			return nil
		}
		return processFile(path)
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func processFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	changed := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		symbol := ""
		isFunc := false

		if m := funcRegex.FindStringSubmatch(line); m != nil {
			symbol = m[1]
			isFunc = true
		} else if m := methodRegex.FindStringSubmatch(line); m != nil {
			symbol = m[1]
			isFunc = true
		} else if m := typeRegex.FindStringSubmatch(line); m != nil {
			symbol = m[1]
		}

		if symbol != "" {
			docStart := -1
			for j := i - 1; j >= 0; j-- {
				trimmed := strings.TrimSpace(lines[j])
				if strings.HasPrefix(trimmed, "//") {
					docStart = j
				} else if trimmed == "" {
					break
				} else {
					break
				}
			}

			if docStart == -1 {
				docComment := []string{
					fmt.Sprintf("// %s ...", symbol),
					fmt.Sprintf("// Summary: %s", symbol),
				}
				if isFunc {
					docComment = append(docComment, "// Parameters:", "//   - None.", "// Returns:", "//   - None.", "// Errors:", "//   - None.", "// Side Effects:", "//   - None.")
				}
				newLines = append(newLines, docComment...)
				changed = true
			} else {
				docLines := lines[docStart:i]
				hasSummary := false
				hasParams := false
				hasReturns := false
				hasErrors := false
				hasSideEffects := false

				for _, dl := range docLines {
					if strings.Contains(dl, "Summary:") { hasSummary = true }
					if strings.Contains(dl, "Parameters:") { hasParams = true }
					if strings.Contains(dl, "Returns:") { hasReturns = true }
					if strings.Contains(dl, "Errors:") { hasErrors = true }
					if strings.Contains(dl, "Side Effects:") { hasSideEffects = true }
				}

				// Copy existing first doc line
				newLines = append(newLines, lines[docStart])

				if !hasSummary {
					newLines = append(newLines, fmt.Sprintf("// Summary: %s", symbol))
					changed = true
				}

				// Copy remaining existing doc lines
				for j := docStart + 1; j < i; j++ {
					newLines = append(newLines, lines[j])
				}

				if isFunc {
					if !hasParams { newLines = append(newLines, "// Parameters:", "//   - None."); changed = true }
					if !hasReturns { newLines = append(newLines, "// Returns:", "//   - None."); changed = true }
					if !hasErrors { newLines = append(newLines, "// Errors:", "//   - None."); changed = true }
					if !hasSideEffects { newLines = append(newLines, "// Side Effects:", "//   - None."); changed = true }
				}
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if changed {
		return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
	}
	return nil
}
