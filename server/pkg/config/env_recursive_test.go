// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpand_RecursiveDefault(t *testing.T) {
	os.Setenv("OTHER_VAR", "expanded_value")
	defer os.Unsetenv("OTHER_VAR")

	// Case 1: Simple recursive default
	input := []byte("Result: ${MISSING_VAR:${OTHER_VAR}}")
	expected := []byte("Result: expanded_value")

	output, err := expand(input)
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(output))

	// Case 2: Recursive default with missing variable (should fail)
	input2 := []byte("Result: ${MISSING_VAR:${ALSO_MISSING}}")
	_, err2 := expand(input2)
	if assert.Error(t, err2) {
		assert.Contains(t, err2.Error(), "missing environment variables")
	}
}
