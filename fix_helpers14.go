package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/integration/e2e_helpers.go")
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "map[string]*configv1.WebsocketCallDefinition{callID: callDef}") {
			lines[i] = strings.Replace(line, "map[string]*configv1.WebsocketCallDefinition{callID: callDef}", "map[string]*configv1.WebsocketCallDefinition{callID: &callDef}", 1)
		}
		if strings.Contains(line, "map[string]*configv1.WebrtcCallDefinition{callID: callDef}") {
			lines[i] = strings.Replace(line, "map[string]*configv1.WebrtcCallDefinition{callID: callDef}", "map[string]*configv1.WebrtcCallDefinition{callID: &callDef}", 1)
		}
		if strings.Contains(line, "map[string]*configv1.MCPCallDefinition{callID: callDef}") {
			lines[i] = strings.Replace(line, "map[string]*configv1.MCPCallDefinition{callID: callDef}", "map[string]*configv1.MCPCallDefinition{callID: &callDef}", 1)
		}
	}
	os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(strings.Join(lines, "\n")), 0644)
}
