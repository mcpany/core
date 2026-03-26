package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("tests/integration/e2e_helpers.go")
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.Contains(line, "callDef := &configv1.HttpCallDefinition_builder{") {
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "}.Build()") {
					lines[j] = strings.Replace(lines[j], "}.Build()", "}", 1)
					break
				}
			}
		}
	}
	os.WriteFile("tests/integration/e2e_helpers.go", []byte(strings.Join(lines, "\n")), 0644)
}
