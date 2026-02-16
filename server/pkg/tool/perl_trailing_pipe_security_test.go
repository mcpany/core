// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPerlTrailingPipeInjection(t *testing.T) {
	// This test attempts to reproduce an RCE vulnerability where Perl open(FH, "cmd|")
	// (trailing pipe) can be injected into a SINGLE-quoted argument string.

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
	}.Build()

	// Perl script: open(FH, '{{input}}') or die "Can't open: $!"; print <FH>;
	// If input is "cmd|", it executes "cmd" and reads its output.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "open(FH, '{{input}}') or die \"Can't open: $!\"; print <FH>;"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("perl_trailing_pipe"),
	}.Build()

	// Use LocalCommandTool to execute locally
	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	f, err := os.CreateTemp("", "rce_check")
	require.NoError(t, err)
	checkFile := f.Name()
	f.Close()
	os.Remove(checkFile) // Ensure it doesn't exist

	// Payload: touch /tmp/xxx|
	payload := "touch " + checkFile + "|"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &ExecutionRequest{
		ToolName:   "perl_trailing_pipe",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err = tool.Execute(ctx, req)

	if err != nil {
		t.Logf("Execution result error: %v", err)
		if strings.Contains(err.Error(), "injection detected") {
			t.Log("Injection detected (PASS)")
			return
		}
	}

	// Check if file exists
	if _, err := os.Stat(checkFile); err == nil {
		t.Fatal("VULNERABILITY CONFIRMED: Perl trailing pipe injection successful (file created)")
	}
}
