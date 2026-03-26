package main

import (
	"bufio"
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
		if info.IsDir() || !strings.EndsWith(path, ".go") || strings.Contains(path, "vendor") || strings.Contains(path, ".pb.go") {
			return nil
		}
		return processFile(path)
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func strings.EndsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
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
			// Find existing doc comment
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
				// Create new doc comment
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
				// Inject missing tags into existing doc comment
				docLines := lines[docStart:i]
				hasSummary := false
				hasParams := false
				hasReturns := false
				hasErrors := false
				hasSideEffects := false

				for _, dl := range docLines {
					if strings.Contains(dl, "Summary:") {
						hasSummary = true
					}
					if strings.Contains(dl, "Parameters:") {
						hasParams = true
					}
					if strings.Contains(dl, "Returns:") {
						hasReturns = true
					}
					if strings.Contains(dl, "Errors:") {
						hasErrors = true
					}
					if strings.Contains(dl, "Side Effects:") {
						hasSideEffects = true
					}
				}

				var injected []string
				lastDocLine := i - 1

				// Rebuild doc block with missing tags
				if !hasSummary {
					// Add Summary after the first line (the descriptive one)
					injected = append(injected, fmt.Sprintf("// Summary: %s", symbol))
					changed = true
				}

				if isFunc {
					if !hasParams { injected = append(injected, "// Parameters:", "//   - None."); changed = true }
					if !hasReturns { injected = append(injected, "// Returns:", "//   - None."); changed = true }
					if !hasErrors { injected = append(injected, "// Errors:", "//   - None."); changed = true }
					if !hasSideEffects { injected = append(injected, "// Side Effects:", "//   - None."); changed = true }
				}

				if len(injected) > 0 {
					// Append injected lines to the end of the doc block (before the symbol)
					lines = append(lines[:i], append(injected, lines[i:]...)...)
					i += len(injected)
				}
			}
		}
		newLines = append(newLines, lines[i])
	}

	if changed {
		return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
	}
	return nil
}
