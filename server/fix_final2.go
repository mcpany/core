package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("tests/integration/e2e_helpers.go")
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.Contains(line, "map[string]*configv1.HttpCallDefinition{callID: &callDef}") {
			lines[i] = strings.Replace(line, "map[string]*configv1.HttpCallDefinition{callID: &callDef}", "map[string]*configv1.HttpCallDefinition{callID: callDef.Build()}", 1)
		}
	}
	os.WriteFile("tests/integration/e2e_helpers.go", []byte(strings.Join(lines, "\n")), 0644)
}
