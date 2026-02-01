// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
	"testing"
)

func TestExpand_DeepRecursion(t *testing.T) {
	// Create a deeply nested variable expansion: ${M:${M:${M:...}}}
	// This exercises the recursion in handleBracedVar -> expand -> handleBracedVar
	depth := 200
	var sb strings.Builder
	for i := 0; i < depth; i++ {
		sb.WriteString("${MISSING:")
	}
	sb.WriteString("foo")
	for i := 0; i < depth; i++ {
		sb.WriteString("}")
	}

	input := []byte(sb.String())

	_, err := expand(input)
	if err == nil {
		t.Fatal("expected error due to recursion depth limit, got nil")
	}

	if !strings.Contains(err.Error(), "recursion depth exceeded") {
		t.Errorf("expected error to contain 'recursion depth exceeded', got: %v", err)
	}
}
