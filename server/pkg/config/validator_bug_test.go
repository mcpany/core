// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"
)

// TestValidateStdioArgs_PythonDashC ...
// Summary: TestValidateStdioArgs_PythonDashC
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Case: python -c "print(1.5)"
	// "print(1.5)" has extension ".5)" so it is treated as a file.
	err := validateStdioArgs(context.Background(), "python", []string{"-c", "print(1.5)"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestValidateStdioArgs_PythonDashC_NoExt ...
// Summary: TestValidateStdioArgs_PythonDashC_NoExt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Case: python -c "print(1)"
	// "print(1)" has no extension. Should be ignored.
	err := validateStdioArgs(context.Background(), "python", []string{"-c", "print(1)"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestValidateStdioArgs_NodeEval ...
// Summary: TestValidateStdioArgs_NodeEval
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// node -e "console.log('hello.world')"
	// "console.log('hello.world')" has extension ".world')"
	err := validateStdioArgs(context.Background(), "node", []string{"-e", "console.log('hello.world')"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestValidateStdioArgs_BashDashC ...
// Summary: TestValidateStdioArgs_BashDashC
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// bash -c "echo hello.world"
	err := validateStdioArgs(context.Background(), "bash", []string{"-c", "echo hello.world"}, "")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
