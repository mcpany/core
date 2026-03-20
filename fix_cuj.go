package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/e2e_sequential/cuj_lifecycle_test.go")
	lines := strings.Split(string(content), "\n")
	var newLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.Contains(line, "time.Sleep(2 * time.Second)") {
		    newLines = append(newLines, "\t\ttime.Sleep(3 * time.Second)")
		    continue
		}

		newLines = append(newLines, line)
	}

	os.WriteFile("server/tests/e2e_sequential/cuj_lifecycle_test.go", []byte(strings.Join(newLines, "\n")), 0644)
}
