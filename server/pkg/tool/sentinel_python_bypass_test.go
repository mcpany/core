// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSentinelRCE_Python_Getattr_Bypass(t *testing.T) {
	// 1. Configure a tool that uses python3 -c with input inside double quotes (eval)
    // This simulates a tool where the developer wants to evaluate a python expression.
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("python3"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "eval(\"{{msg}}\")"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("msg"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("python_getattr_bypass")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"python_getattr_bypass",
	)

	// 2. Craft a malicious input that bypasses static analysis
	// - Uses string concatenation to hide "__import__"
	// - Uses getattr to hide "system"
	// - Uses __builtins__ to access __import__
    // - Uses single quotes (allowed in Double Quoted context)
	payload := "getattr(__builtins__['__im'+'port__']('os'), 'sys'+'tem')('id')"

	inputMap := map[string]interface{}{
		"msg": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "python_getattr_bypass",
		ToolInputs: inputBytes,
	}

	// 3. Execute
	_, err := tool.Execute(context.Background(), req)

	// 4. Assert
	if err == nil {
		t.Log("Vulnerability confirmed: Payload was accepted!")
		t.Fail()
	} else {
        t.Logf("Blocked with error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Payload should be blocked")
	}
}
