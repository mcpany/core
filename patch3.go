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

	// Fix the unused callDef error 1
	str = strings.Replace(str, `	callDef := configv1.HttpCallDefinition_builder{
		Endpoint: "http://example.com",
		Method:   "GET",
	}.Build()`, ``, 1)

	// Fix the unused callDef error 2
	str = strings.Replace(str, `	callDef := configv1.HttpCallDefinition_builder{
		Endpoint: "http://example.com/status",
		Method:   "GET",
	}.Build()`, ``, 1)

	err = os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched successfully")
}
