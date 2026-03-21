package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("server/tests/integration/e2e_helpers.go")
	if err != nil {
		panic(err)
	}

	modified := strings.ReplaceAll(string(content), "var roots []string\n\tfor _, base := range []string{os.Getenv(\"TEST_SRCDIR\"), os.Getenv(\"RUNFILES_DIR\")} {", "roots := make([]string, 0, 2)\n\tfor _, base := range []string{os.Getenv(\"TEST_SRCDIR\"), os.Getenv(\"RUNFILES_DIR\")} {")

	modified = strings.ReplaceAll(modified, "func symlinkIfPresent(src, dst string) error {\n\tif _, err := os.Stat(src); err != nil {\n\t\treturn nil\n\t}", "func symlinkIfPresent(src, dst string) error {\n\tif _, err := os.Stat(src); err != nil {\n\t\treturn nil //nolint:nilerr // Ignore if source doesn't exist\n\t}")

	err = os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(modified), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched successfully")
}
