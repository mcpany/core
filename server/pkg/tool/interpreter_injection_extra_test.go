// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestPHPInjectionRepro(t *testing.T) {
	// Setup a Shell wrapping PHP tool config to emulate quoted context
	// Command: sh -c "php -r '{{code}}'"

	properties := map[string]*structpb.Value{
		"code": structpb.NewStringValue(""),
	}

	toolProto := pb.Tool_builder{
		Name: proto.String("php-exec"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: properties,
				}),
			},
		},
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"default": configv1.CommandLineCallDefinition_builder{
				Args: []string{"-c", "php -r '{{code}}'"},
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
					}.Build(),
				},
			}.Build(),
		},
	}.Build()

	callDef := service.GetCalls()["default"]

	cmdTool := NewLocalCommandTool(toolProto, service, callDef, nil, "default")

	// Test cases for missing dangerous functions
	testCases := []struct {
		name      string
		inputCode string
		shouldFail bool // If true, we expect validation error (post-fix). Initially we expect false (pre-fix).
	}{
		{
			name:      "passthru",
			inputCode: "passthru(\"echo vulnerable\");", // Use double quotes to avoid shell single-quote breakout
			shouldFail: true, // Now fixed, should detect
		},
		{
			name:      "include",
			inputCode: "include \"/etc/passwd\";", // include is a construct, no parens needed (but here used with quotes)
			shouldFail: true, // Now fixed
		},
		{
			name:      "assert",
			inputCode: "assert(\"phpinfo()\");", // use safe inner payload to avoid triggering 'system' check
			shouldFail: true, // Now fixed
		},
		{
			name:      "dl",
			inputCode: "dl(\"extension.so\");",
			shouldFail: true, // Now fixed
		},
		{
			name:      "case_insensitive_bypass_parens",
			inputCode: "PaSsThRu(\"echo vulnerable\");",
			shouldFail: true, // Detected by cleanVal check
		},
		{
			name:      "case_insensitive_bypass_no_parens",
			inputCode: "InClUdE \"/etc/passwd\";",
			shouldFail: true, // Should be detected by updated checkUnquotedKeywords
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName:   "php-exec",
				ToolInputs: []byte(`{"code": "` + strings.ReplaceAll(tc.inputCode, "\"", "\\\"") + `"}`),
				DryRun:     true, // Use DryRun to trigger validation without execution
			}

			_, err := cmdTool.Execute(context.Background(), req)

			if tc.shouldFail {
				assert.Error(t, err, "Expected validation error for %s", tc.name)
				if err != nil {
					assert.Contains(t, err.Error(), "injection detected", "Error should mention injection detected")
				}
			} else {
				// Pre-fix: We expect NO error for missing keywords
				assert.NoError(t, err, "Expected no error for %s (vulnerability reproduction)", tc.name)
			}
		})
	}
}

func TestRubyInjectionRepro(t *testing.T) {
	// Setup a Shell wrapping Ruby tool config
	// Command: sh -c "ruby -e '{{code}}'"

	properties := map[string]*structpb.Value{
		"code": structpb.NewStringValue(""),
	}

	toolProto := pb.Tool_builder{
		Name: proto.String("ruby-exec"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: properties,
				}),
			},
		},
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"default": configv1.CommandLineCallDefinition_builder{
				Args: []string{"-c", "ruby -e '{{code}}'"},
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{Name: proto.String("code")}.Build(),
					}.Build(),
				},
			}.Build(),
		},
	}.Build()

	callDef := service.GetCalls()["default"]

	cmdTool := NewLocalCommandTool(toolProto, service, callDef, nil, "default")

	// Test cases for missing dangerous functions
	testCases := []struct {
		name      string
		inputCode string
		shouldFail bool
	}{
		{
			name:      "syscall",
			inputCode: "syscall(\"echo vulnerable\")",
			shouldFail: true, // Now fixed
		},
		{
			name:      "load",
			inputCode: "load(\"vulnerable.rb\")",
			shouldFail: true, // Now fixed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName:   "ruby-exec",
				ToolInputs: []byte(`{"code": "` + strings.ReplaceAll(tc.inputCode, "\"", "\\\"") + `"}`),
				DryRun:     true,
			}

			_, err := cmdTool.Execute(context.Background(), req)

			if tc.shouldFail {
				assert.Error(t, err, "Expected validation error for %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for %s (vulnerability reproduction)", tc.name)
			}
		})
	}
}
