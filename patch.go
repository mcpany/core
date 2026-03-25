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

	str = strings.Replace(str, `	var roots []string`, `	roots := make([]string, 0, 4)`, 1)

	err = os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched successfully")
}
