package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/integration/e2e_helpers.go")
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "var roots []string") {
			lines[i] = "\troots := make([]string, 0, 4)"
		}
		if strings.Contains(line, "return nil") && i == 272 {
		    // Check line 271 for if _, err := os.Stat(src); err != nil {
		    if strings.Contains(lines[i-1], "if _, err := os.Stat(src); err != nil {") {
			    lines[i] = "\t\treturn nil //nolint:nilerr // Ignore if source doesn't exist"
			}
		}
	}
	os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(strings.Join(lines, "\n")), 0644)
}
