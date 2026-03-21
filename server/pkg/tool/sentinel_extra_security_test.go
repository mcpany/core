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

func TestSentinel_GitArgumentInjection(t *testing.T) {
	// 1. Configure a tool that uses git with user input in an argument
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("git"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"fetch", "origin", "--upload-pack={{pack}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("pack"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("git_fetch")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"git_fetch",
	)

	// 2. Craft a malicious input that attempts to inject a flag via space
	// If the argument splitting is vulnerable, this would be interpreted as:
	// --upload-pack=foo -c core.sshCommand=echo pwned
	payload := "foo -c core.sshCommand=echo pwned"

	inputMap := map[string]interface{}{
		"pack": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "git_fetch",
		ToolInputs: inputBytes,
	}

	// 3. Execute
	// We expect this to FAIL because git will try to find an upload-pack named "foo -c core.sshCommand=echo pwned"
	// and fail to execute it (or fail to connect).
	// Crucially, it should NOT execute the injected command.
	// Since we don't have a real git repo, it will fail with exit code 128 or similar.
	// The important thing is that `exec.Command` treats it as a single argument.
	// We can't easily verify that the *command* wasn't executed without checking side effects or mocking executor.
	// But `LocalCommandTool` logic ensures that for non-shell commands, args are passed directly.

	// However, we want to ensure `LocalCommandTool` itself doesn't block this input if it considers it safe.
	// It should pass the "SafePath" check (no - at start) and since it's not a shell,
	// `checkForShellInjection` is skipped.
	// So `Execute` should proceed to call `executor.Execute`.

	res, err := tool.Execute(context.Background(), req)

	// 4. Assert
	// We check that the argument is passed as a SINGLE string in the result args if execution succeeded (git might just warn and exit 0).
	if res != nil {
		if resultArgs, ok := res.(map[string]interface{})["args"].([]string); ok {
			found := false
			for _, arg := range resultArgs {
				if arg == "--upload-pack="+payload {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected payload to be passed as a single argument, got: %v", resultArgs)
		}
	} else if err != nil {
		// If it failed to execute (e.g. exit code != 0), that's also fine as long as it's not due to our injection check blocking it
		assert.NotContains(t, err.Error(), "injection detected")
	}
}

func TestSentinel_GitExtProtocol(t *testing.T) {
	svc := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("git"),
		WorkingDirectory: proto.String("."),
	}.Build()

	stringType := configv1.ParameterType(configv1.ParameterType_value["STRING"])

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
					Type: &stringType,
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := &v1.Tool{}
	toolProto.SetName("git_clone")

	tool := NewLocalCommandTool(
		toolProto,
		svc,
		callDef,
		nil,
		"git_clone",
	)

	payload := "ext::sh -c echo pwned% "

	inputMap := map[string]interface{}{
		"url": payload,
	}
	inputBytes, _ := json.Marshal(inputMap)

	req := &ExecutionRequest{
		ToolName:   "git_clone",
		ToolInputs: inputBytes,
	}

	res, err := tool.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	// We now catch this earlier with the dangerous scheme check
	// "dangerous scheme detected: ext" OR "git ext:: protocol is not allowed"
	if err != nil {
		isExt := assert.Contains(t, err.Error(), "dangerous scheme detected: ext")
		if !isExt {
			assert.Contains(t, err.Error(), "git ext:: protocol is not allowed")
		}
	}
}
