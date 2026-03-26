package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("tests/integration/e2e_helpers.go")
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.Contains(line, "map[string]*configv1.HttpCallDefinition{callID: callDef.Build()}") {
			lines[i] = strings.Replace(line, "map[string]*configv1.HttpCallDefinition{callID: callDef.Build()}", "map[string]*configv1.HttpCallDefinition{callID: callDef}", 1)
		}
		if strings.Contains(line, "callDef := configv1.HttpCallDefinition_builder{") {
			lines[i] = strings.Replace(line, "callDef := configv1.HttpCallDefinition_builder{", "callDef := &configv1.HttpCallDefinition_builder{", 1)
		}
	}
	os.WriteFile("tests/integration/e2e_helpers.go", []byte(strings.Join(lines, "\n")), 0644)
}
