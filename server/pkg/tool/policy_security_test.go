package tool

import (
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPolicyEvasion_DuplicateKeys(t *testing.T) {
	// Scenario: Policy allows requests containing "safe_value", defaulting to DENY.
	// Attacker sends JSON with duplicate keys: first one is "safe_value", second is "dangerous_value".
	// Regex matches "safe_value" -> ALLOW.
	// JSON Parser (standard Go) takes the last value -> "dangerous_value".

	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_DENY.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				ArgumentRegex: proto.String(`.*"safe_value".*`),
				Action:        configv1.CallPolicy_ALLOW.Enum(),
			}.Build(),
		},
	}.Build()

	compiled, err := CompileCallPolicies([]*configv1.CallPolicy{policy})
	assert.NoError(t, err)

	// JSON with duplicate keys
	payload := []byte(`{"param": "safe_value", "param": "dangerous_value"}`)

	// 1. Verify Policy Evaluation
	allowed, err := EvaluateCompiledCallPolicy(compiled, "tool", "call", payload)
	assert.NoError(t, err)
	// If this is false, the policy blocked the request because normalized JSON "dangerous_value" doesn't match "safe_value"
	assert.False(t, allowed, "Policy should DENY the request because normalized JSON resolved to 'dangerous_value'")

	// 2. Verify JSON Parsing
	var parsed map[string]interface{}
	err = json.Unmarshal(payload, &parsed)
	assert.NoError(t, err)

	val := parsed["param"]
	assert.Equal(t, "dangerous_value", val, "JSON parser should resolve to the last key value")
}
