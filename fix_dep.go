package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/integration/deployment_test.go")
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "t.Skip") {
		    lines[i] = "\t// " + line
		}
	}
	os.WriteFile("server/tests/integration/deployment_test.go", []byte(strings.Join(lines, "\n")), 0644)
}
