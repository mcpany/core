// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintResults(t *testing.T) {
	results := []CheckResult{
		{ServiceName: "ServiceA", Status: StatusOk, Message: "All good"},
		{ServiceName: "ServiceB", Status: StatusWarning, Message: "Something minor"},
		{ServiceName: "ServiceC", Status: StatusError, Message: "Critical failure"},
		{ServiceName: "ServiceD", Status: StatusSkipped, Message: "Disabled"},
	}

	var buf bytes.Buffer
	PrintResults(&buf, results)

	output := buf.String()

	if !strings.Contains(output, "✅") {
		t.Errorf("Expected OK icon")
	}
	if !strings.Contains(output, "⚠️") {
		t.Errorf("Expected Warning icon")
	}
	if !strings.Contains(output, "❌") {
		t.Errorf("Expected Error icon")
	}
	if !strings.Contains(output, "⏭️") {
		t.Errorf("Expected Skipped icon")
	}
	if !strings.Contains(output, "ServiceA") {
		t.Errorf("Expected ServiceA name")
	}
	if !strings.Contains(output, "Critical failure") {
		t.Errorf("Expected error message")
	}
}
