// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyEnvVarsFromSlice_Determinism(t *testing.T) {
	// Scenario: Conflicting environment variables.
	// MCPANY__A=val sets "a" to a string "val".
	// MCPANY__A__B=val2 sets "a" to a map {"b": "val2"}.

	env1 := []string{"MCPANY__A=val", "MCPANY__A__B=val2"}
	env2 := []string{"MCPANY__A__B=val2", "MCPANY__A=val"}

	// Note on expected behavior:
	// We sort the environment variables before processing.
	// "MCPANY__A" < "MCPANY__A__B" lexicographically.
	// So "MCPANY__A" is processed first, setting a="val".
	// Then "MCPANY__A__B" is processed, overwriting a={"b": "val2"}.
	//
	// So the expected result is {"a": {"b": "val2"}}.

	expected := map[string]interface{}{
		"a": map[string]interface{}{"b": "val2"},
	}

	// Case 1: Already sorted input
	m1 := make(map[string]interface{})
	applyEnvVarsFromSlice(m1, env1, nil)
	assert.Equal(t, expected, m1, "Expected m1 to match sorted result")

	// Case 2: Reverse sorted input
	m2 := make(map[string]interface{})
	applyEnvVarsFromSlice(m2, env2, nil)
	assert.Equal(t, expected, m2, "Expected m2 to match sorted result (A processed before A__B)")

	// Final verification: determinism
	assert.Equal(t, m1, m2, "Expected consistent result regardless of input order")
}
