// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"
)

func TestValidateStdioArgs_PythonDashC(t *testing.T) {
	// Case: python -c "print(1.5)"
	// "print(1.5)" has extension ".5)" so it is treated as a file.
	err := validateStdioArgs(context.Background(), "python", []string{"-c", "print(1.5)"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateStdioArgs_PythonDashC_NoExt(t *testing.T) {
	// Case: python -c "print(1)"
	// "print(1)" has no extension. Should be ignored.
	err := validateStdioArgs(context.Background(), "python", []string{"-c", "print(1)"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateStdioArgs_NodeEval(t *testing.T) {
	// node -e "console.log('hello.world')"
	// "console.log('hello.world')" has extension ".world')"
	err := validateStdioArgs(context.Background(), "node", []string{"-e", "console.log('hello.world')"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateStdioArgs_BashDashC(t *testing.T) {
	// bash -c "echo hello.world"
	err := validateStdioArgs(context.Background(), "bash", []string{"-c", "echo hello.world"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
