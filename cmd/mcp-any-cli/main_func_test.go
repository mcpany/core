package main

import (
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	// Keep a reference to the original os.Args
	oldArgs := os.Args
	// Defer the restoration of the original os.Args
	defer func() { os.Args = oldArgs }()

	// Set the arguments for the test
	os.Args = []string{"mcp-any-cli", "version"}

	// Call the execute function
	err := execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Set the arguments for the test to cover the error path
	os.Args = []string{"mcp-any-cli", "unknown"}

	// Call the execute function
	err = execute()
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}

func TestMain(t *testing.T) {
	// Keep a reference to the original os.Args
	oldArgs := os.Args
	// Defer the restoration of the original os.Args
	defer func() { os.Args = oldArgs }()

	// Set the arguments for the test
	os.Args = []string{"mcp-any-cli", "version"}

	// Call the main function
	main()
}
