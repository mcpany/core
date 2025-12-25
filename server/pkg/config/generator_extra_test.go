// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bufio"
	"bytes"
	"testing"
)

func TestGenerator_Prompt_Bug(t *testing.T) {
	// Input without a trailing newline
	input := "some input"
	reader := bufio.NewReader(bytes.NewBufferString(input))
	g := &Generator{
		Reader: reader,
	}

	result, err := g.prompt("prompt: ")

	// We expect "some input" but the current implementation returns "" and an error (probably EOF)
	// We want to handle EOF gracefully if there is content.

	if result != "some input" {
		t.Errorf("Expected 'some input', got '%s'", result)
	}

	// The current implementation returns an error on EOF even if there is data.
	// Depending on how we want to define "bug", usually tools should accept the last line even without newline.
	if err != nil {
		// It returns error because of EOF, but we might want to check if it returned the content at least?
		// Current code: return "", err
		t.Logf("Got error as expected from current implementation: %v", err)
	}
}
