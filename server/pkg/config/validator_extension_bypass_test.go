// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"
)

func TestValidateStdioArgs_Bypass_NoExtension(t *testing.T) {
	ctx := context.Background()

	// Case 1: File with extension (Should be checked)
	// "/restricted/script.py" looks like a file (ext + abs path).
	err := validateStdioArgs(ctx, "python", []string{"/restricted/script.py"}, "")
	if err == nil {
		t.Errorf("script.py check skipped (unexpected)")
	} else {
		t.Logf("script.py check triggered: %v", err)
	}

	// Case 2: File without extension BUT is a path (Should be checked)
	// "/restricted/script_no_ext" is absolute path.
	err = validateStdioArgs(ctx, "python", []string{"/restricted/script_no_ext"}, "")
	if err == nil {
		t.Errorf("script_no_ext (abs path) check SKIPPED (Vulnerability still exists)")
	} else {
		t.Logf("script_no_ext (abs path) check triggered (Fix confirmed): %v", err)
	}

	// Case 3: Relative path (Should be checked)
	// "./script" has separator.
	err = validateStdioArgs(ctx, "python", []string{"./script_no_ext"}, "")
	if err == nil {
		t.Errorf("./script_no_ext check SKIPPED (Vulnerability still exists)")
	} else {
		t.Logf("./script_no_ext check triggered (Fix confirmed): %v", err)
	}

	// Case 4: Bare word / Subcommand (Should be SKIPPED/VALID)
	// "run" or "status". No ext, no separators.
	err = validateStdioArgs(ctx, "go", []string{"run"}, "")
	if err != nil {
		t.Errorf("bare word 'run' was CHECKED (Regression!): %v", err)
	} else {
		t.Logf("bare word 'run' skipped (Usability preserved)")
	}
}
