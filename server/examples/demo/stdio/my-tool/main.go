// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a demo tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Request represents a tool request.
type Request struct {
	Name string `json:"name"`
}

// Response represents a tool response.
type Response struct {
	Message string `json:"message"`
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var req Request
	if err := json.Unmarshal(input, &req); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling JSON: %v\n", err)
		os.Exit(1)
	}

	resp := Response{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}

	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}
}
