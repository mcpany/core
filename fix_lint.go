package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/e2e_sequential/cuj_lifecycle_test.go")
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "useLocal := true // Force local mode") {
			lines[i] = "\tuseLocal := true // Force local mode\n\t_ = useLocal"
		}
	}
	os.WriteFile("server/tests/e2e_sequential/cuj_lifecycle_test.go", []byte(strings.Join(lines, "\n")), 0644)
}
