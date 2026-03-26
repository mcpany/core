package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	files := []string{"server/tests/framework/gemini.go", "server/tests/framework/claude.go"}
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", path, err)
			continue
		}
		lines := strings.Split(string(content), "\n")
		var newLines []string
		for _, line := range lines {
			if strings.Contains(line, "DefaultModel =") || strings.Contains(line, "DefaultClaudeModel =") {
				newLines = append(newLines, "// Summary: Default model constant.")
			}
			newLines = append(newLines, line)
		}
		os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
	}
}
