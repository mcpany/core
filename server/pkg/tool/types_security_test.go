// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestSecurityHardening_SpaceInArguments verifies that spaces are allowed in arguments for command tools.
// This is a regression test for a previous security update that blocked spaces to prevent argument injection,
// which inadvertently broke usability for tools like git.
func TestSecurityHardening_SpaceInArguments(t *testing.T) {
	t.Parallel()
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"commit", "-m", "{{msg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("msg")}},
		},
	}
	// "git" is in the shell command list, triggering strict checks
	cmdTool := newCommandTool("git", callDef)

	inputData := map[string]interface{}{"msg": "initial commit"}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = cmdTool.Execute(context.Background(), req)

	// Expectation: Should NOT fail with "shell injection detected".
	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected", "Space in argument should be allowed")
	}
}

// TestSecurityHardening_SSRF_Curl verifies that network tools like curl enforce egress policies
// by checking arguments for unsafe URLs.
func TestSecurityHardening_SSRF_Curl(t *testing.T) {
	// Since TestMain mocks IsSafeURL to be permissive, we must mock it to be strict for this test.
	// We want to verify that CommandTool *calls* validation.IsSafeURL.
	originalValidator := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalValidator }()

	validation.IsSafeURL = func(urlStr string) error {
		if strings.Contains(urlStr, "169.254.169.254") {
			return fmt.Errorf("mocked unsafe url error")
		}
		return nil
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("url")}},
		},
	}
	// "curl" is in the shell command list and network command list
	cmdTool := newCommandTool("curl", callDef)

	// Use an unsafe IP (Link-Local / Metadata service)
	inputData := map[string]interface{}{"url": "http://169.254.169.254/latest/meta-data/"}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = cmdTool.Execute(context.Background(), req)

	// Expectation: Should fail with "unsafe url".
	require.Error(t, err, "Should fail with unsafe url error")
	assert.Contains(t, err.Error(), "unsafe url", "Error message should mention unsafe url")
}
