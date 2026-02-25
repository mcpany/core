// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestPythonRCEBypass(t *testing.T) {
	// Target: Python Code Execution via __getattribute__
	// Constraint: 'os' keyword is blocked as object, but if we can get a reference to it,
	// we can call __getattribute__('system')('ls').

	// Bypass using semicolon and assignment to get reference to os
	val3 := `x=os; x.__getattribute__('system')('ls')`

	err := checkInterpreterFunctionCalls(val3, "python")
	if err != nil {
		t.Logf("Blocked Python RCE: %v", err)
	} else {
		t.Errorf("Bypass: Python RCE accepted: %s", val3)
	}
}
