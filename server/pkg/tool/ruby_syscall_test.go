// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RubySyscallInjection(t *testing.T) {
	// This test reproduces an RCE vulnerability where Ruby 'syscall'
	// can be executed without quotes or parens.

	// Use syscall 20 (getpid) as a harmless RCE proof
	// Payload: syscall 20
	payload := "syscall 20"

	tool := v1.Tool_builder{
		Name: proto.String("ruby-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// Ruby script: puts #{input}
	// Or just execute input as code: ruby -e {{input}}
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "ruby-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// We expect the execution to be BLOCKED by security checks.
	// If it succeeds (err == nil), it means vulnerability exists.
	if err == nil {
		t.Fatal("VULNERABILITY CONFIRMED: Ruby syscall execution allowed")
	}

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "interpreter injection detected"), "Expected interpreter injection error, got: %v", err)
	assert.True(t, strings.Contains(err.Error(), "syscall"), "Expected 'syscall' to be blocked")
}

func TestLocalCommandTool_RubySendInjection(t *testing.T) {
	// This test checks if 'send' (method invocation) is blocked in Ruby.
	// Payload: send(:eval, 'puts 1')
	// Since parens are blocked in unquoted context, we can try without parens?
	// send :eval, "puts 1" -> blocked by " (quote)
	// send :eval, 1 -> harmless but proves 'send' is allowed

	// But wait, checkUnquotedInjection blocks : ? No.
	// It blocks ".
	// So we cannot pass string "puts 1".
	// But we can construct strings using char codes.
	// 49.chr -> "1".
	// puts 1.

	// Payload: send :system, 108.chr+115.chr (ls)
	// system is blocked by function keywords?
	// send is NOT blocked.

	// So if I can construct "ls" without quotes...
	// 108.chr + 115.chr -> "ls"
	// send :system, ...
	// Wait, :system is a symbol. :system contains 'system'.
	// Is :system blocked?
	// checkUnquotedKeywords checks 'system'.
	// It tokenizes by delimiters. : is delimiter?
	// No, : is not delimiter for checkUnquotedKeywords (it uses IsWordChar).
	// : is not WordChar.
	// So :system -> : and system.
	// So 'system' is checked. And it is blocked!

	// So I cannot use :system.
	// Can I use "system"? No quotes.
	// Can I use string construction for method name?
	// send 115.chr+121.chr+... (system), args...

	// So yes, I can invoke system via send!

	// Payload: send 115.chr+121.chr+115.chr+116.chr+101.chr+109.chr, 108.chr+115.chr
	// (send "system", "ls")
	// "system" string construction:
	// s=115, y=121, s=115, t=116, e=101, m=109
	// l=108, s=115

	payload := "send 115.chr+121.chr+115.chr+116.chr+101.chr+109.chr, 108.chr+115.chr"

	tool := v1.Tool_builder{
		Name: proto.String("ruby-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "ruby-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Fatal("VULNERABILITY CONFIRMED: Ruby send execution allowed")
	}

	assert.Error(t, err)
	// We expect 'send' to be blocked now
	assert.True(t, strings.Contains(err.Error(), "interpreter injection detected"), "Expected interpreter injection error, got: %v", err)
	assert.True(t, strings.Contains(err.Error(), "send"), "Expected 'send' to be blocked")
}
