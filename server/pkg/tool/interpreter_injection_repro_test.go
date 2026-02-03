// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"
)

func TestInterpreterInjection_Repro(t *testing.T) {
	// Awk Injection via pipe
	t.Run("awk_pipe_injection", func(t *testing.T) {
		cmd := "awk"
		// awk '{{input}}'
		tool := createTestCommandToolWithTemplate(cmd, "'{{input}}'")

		// Payload: BEGIN { print "pwned" | "sh" }
		// This uses single quotes for the tool argument definition, so we are in quoteLevel 2.
		// quoteLevel 2 blocks ' and ` and system(.
		// But it doesn't block | or "sh".
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "BEGIN { print \"pwned\" | \"sh\" }"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// Expectation: Currently this passes (err == nil) because validation is insufficient.
		if err == nil {
			t.Fatal("Expected error but got nil (Vulnerability present: validation passed)")
		}
		if !strings.Contains(err.Error(), "injection detected") {
			t.Fatalf("Expected shell injection error, got: %v", err)
		}
	})

	// Perl Injection via open
	t.Run("perl_open_injection", func(t *testing.T) {
		cmd := "perl"
		// perl -e '{{input}}'
		tool := createTestCommandToolWithTemplate(cmd, "-e '{{input}}'")

		// Payload: open(F, "| sh")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "open(F, \"| sh\");"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if err == nil {
			t.Fatal("Expected error but got nil (Vulnerability present: validation passed)")
		}
		if !strings.Contains(err.Error(), "injection detected") {
			t.Fatalf("Expected shell injection error, got: %v", err)
		}
	})

	// PHP Injection via passthru
	t.Run("php_passthru_injection", func(t *testing.T) {
		cmd := "php"
		// php -r '{{input}}'
		tool := createTestCommandToolWithTemplate(cmd, "-r '{{input}}'")

		// Payload: passthru("ls");
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "passthru(\"ls\");"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if err == nil {
			t.Fatal("Expected error but got nil (Vulnerability present: validation passed)")
		}
		if !strings.Contains(err.Error(), "injection detected") {
			t.Fatalf("Expected shell injection error, got: %v", err)
		}
	})
}
