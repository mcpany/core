// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestEvaluateCompiledCallPolicy_Coverage(t *testing.T) {
	allow := configv1.CallPolicy_ALLOW
	deny := configv1.CallPolicy_DENY

	policies := []*configv1.CallPolicy{
		{
			DefaultAction: &deny,
			Rules: []*configv1.CallPolicyRule{
				{
					Action:        &allow,
					NameRegex:     proto.String("match.*"),
					CallIdRegex:   proto.String("id.*"),
					ArgumentRegex: proto.String(".*secret.*"),
				},
			},
		},
	}

	compiled, err := CompileCallPolicies(policies)
	assert.NoError(t, err)

	// Case 1: Name mismatch
	allowed, err := EvaluateCompiledCallPolicy(compiled, "nomatch", "id1", json.RawMessage(`{"a":"secret"}`))
	assert.NoError(t, err)
	assert.False(t, allowed) // Default DENY

	// Case 2: Call ID mismatch
	allowed, err = EvaluateCompiledCallPolicy(compiled, "match1", "noid", json.RawMessage(`{"a":"secret"}`))
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Case 3: Arg mismatch
	allowed, err = EvaluateCompiledCallPolicy(compiled, "match1", "id1", json.RawMessage(`{"a":"public"}`))
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Case 4: All match -> ALLOW
	allowed, err = EvaluateCompiledCallPolicy(compiled, "match1", "id1", json.RawMessage(`{"a":"secret"}`))
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCompileCallPolicies_InvalidRegex_Extra(t *testing.T) {
	policies := []*configv1.CallPolicy{
		{
			Rules: []*configv1.CallPolicyRule{
				{NameRegex: proto.String("[")},
			},
		},
	}
	_, err := CompileCallPolicies(policies)
	assert.Error(t, err)

	policies = []*configv1.CallPolicy{
		{
			Rules: []*configv1.CallPolicyRule{
				{CallIdRegex: proto.String("[")},
			},
		},
	}
	_, err = CompileCallPolicies(policies)
	assert.Error(t, err)

	policies = []*configv1.CallPolicy{
		{
			Rules: []*configv1.CallPolicyRule{
				{ArgumentRegex: proto.String("[")},
			},
		},
	}
	_, err = CompileCallPolicies(policies)
	assert.Error(t, err)
}
