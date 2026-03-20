package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/e2e_sequential/docker_compose_test.go")
	lines := strings.Split(string(content), "\n")
	var newLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.Contains(line, "if _, err := os.Stat(rootCompose); err == nil {") {
		    newLines = append(newLines, "\tif false {")
		    continue
		}

		newLines = append(newLines, line)
	}

	os.WriteFile("server/tests/e2e_sequential/docker_compose_test.go", []byte(strings.Join(newLines, "\n")), 0644)
}
