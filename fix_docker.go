package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/e2e_sequential/docker_compose_test.go")
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "if false {") {
			lines[i] = "\t// _ = originalCompose"
		}
	}
	os.WriteFile("server/tests/e2e_sequential/docker_compose_test.go", []byte(strings.Join(lines, "\n")), 0644)
}
