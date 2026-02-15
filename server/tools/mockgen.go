//go:build generate
// +build generate

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	generateMocks()
}

func generateMocks() {
	fmt.Println("Generating mocks...")
	cmd := exec.Command("go", "generate", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating mocks: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Mocks generated successfully.")
}
