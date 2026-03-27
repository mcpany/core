package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile("pkg/app/server.go")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	content := string(b)

	// Just use `_ = err` to bypass `errcheck` for all unhandled errors!
	// Some linters complain about `if err := ...; err != nil {}` if they want you to do something with the error, but we're just logging it.
	// Wait, we didn't add the error checking initially! In a1fd5a1 we DID NOT add error checking!
	// The linter was FAILING on a1fd5a1!
	// The problem is something we did in a1fd5a1!
	// What did we do in a1fd5a1?

	// We removed math/rand and added crypto/rand.
	// We changed `toTrace` signature.

	fmt.Println("Done")
}
