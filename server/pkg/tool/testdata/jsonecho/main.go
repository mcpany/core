// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"io"
	"os"
)

// main executes the json echo test tool.
//
// Summary: Executes json echo tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - Exits with 1 on failure.
//
// Side Effects:
//   - Reads from stdin, writes to stdout.
func main() {
	var data map[string]interface{}
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		if err != io.EOF {
			os.Exit(1)
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
		os.Exit(1)
	}
}
