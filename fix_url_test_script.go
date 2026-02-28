package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile("server/pkg/validation/url_test.go")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	content := string(b)

	if !strings.Contains(content, "\"strings\"") {
		content = strings.Replace(content, "import (", "import (\n\t\"strings\"", 1)
	}

	// Remove the hacky functions
	idx := strings.Index(content, "func containsStr")
	if idx != -1 {
		content = content[:idx]
	}

	// Replace containsStr with strings.Contains
	content = strings.ReplaceAll(content, "containsStr(err.Error(), tt.errMsg)", "strings.Contains(err.Error(), tt.errMsg)")

	err = os.WriteFile("server/pkg/validation/url_test.go", []byte(content), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Done")
}
