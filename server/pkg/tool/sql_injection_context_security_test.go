// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSQLInjectionDoubleQuote(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("psql"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "\"SELECT * FROM users WHERE id = {{input}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputProperties, _ := structpb.NewStruct(map[string]interface{}{
		"input": map[string]interface{}{"type": "string"},
	})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":       structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(inputProperties),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        proto.String("sql_injection"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	payload := "1; DROP TABLE users; --"

	req := &ExecutionRequest{
		ToolName:   "sql_injection",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for SQL injection in double quotes")
		t.Fail()
	} else {
		t.Logf("Got expected error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}

func TestSQLAllowedSemicolonInQuotes(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("psql"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "\"INSERT INTO t VALUES ('{{input}}')\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputProperties, _ := structpb.NewStruct(map[string]interface{}{
		"input": map[string]interface{}{"type": "string"},
	})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":       structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(inputProperties),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        proto.String("sql_safe_semicolon"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	payload := "value; with semicolon"

	req := &ExecutionRequest{
		ToolName:   "sql_safe_semicolon",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
	} else {
		assert.NotContains(t, err.Error(), "injection detected", "Validation should allow semicolon in quoted string")
		assert.Contains(t, err.Error(), "executable file not found", "Expected execution error")
	}
}

func TestSQLQuoteBreakout(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("psql"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "\"SELECT ... name='{{input}}'\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputProperties, _ := structpb.NewStruct(map[string]interface{}{
		"input": map[string]interface{}{"type": "string"},
	})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":       structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(inputProperties),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        proto.String("sql_quote_breakout"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	payload := "foo' OR 1=1 --"

	req := &ExecutionRequest{
		ToolName:   "sql_quote_breakout",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for SQL quote breakout")
		t.Fail()
	} else {
		t.Logf("Got expected error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
		assert.Contains(t, err.Error(), "single quote", "Expected single quote detection")
	}
}

func TestSQLBackslashInjection(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("psql"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "\"SELECT ... name='{{input}}'\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputProperties, _ := structpb.NewStruct(map[string]interface{}{
		"input": map[string]interface{}{"type": "string"},
	})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":       structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(inputProperties),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        proto.String("sql_backslash"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	// Payload: \
	// JSON string must escape backslash: "\\" represents \
	// Go string literal for "\\" is "\\\\"
	payload := "\\\\"

	req := &ExecutionRequest{
		ToolName:   "sql_backslash",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for SQL backslash injection")
		t.Fail()
	} else {
		t.Logf("Got expected error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
		assert.Contains(t, err.Error(), "backslash", "Expected backslash detection")
	}
}
