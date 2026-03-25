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

	str := string(content)

	str = strings.Replace(str, `func symlinkIfPresent(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	return os.Symlink(src, dst)
}`, `func symlinkIfPresent(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		// Ignore if the source directory does not exist.
		return nil //nolint:nilerr
	}
	return os.Symlink(src, dst)
}`, 1)

	err = os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched successfully")
}
