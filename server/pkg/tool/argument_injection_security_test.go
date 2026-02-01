// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForArgumentInjection_Plus(t *testing.T) {
	// vim +:r/etc/passwd
	err := checkForArgumentInjection("+:r/etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "argument injection detected: value starts with '-' or '+'")
}

func TestCheckForArgumentInjection_PlusNumber(t *testing.T) {
	// tail +10
	err := checkForArgumentInjection("+10")
	assert.NoError(t, err, "Should allow positive numbers")

	// tail +10.5
	err = checkForArgumentInjection("+10.5")
	assert.NoError(t, err, "Should allow positive floats")
}

func TestCheckForArgumentInjection_Minus(t *testing.T) {
	// ls -la
	err := checkForArgumentInjection("-la")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "argument injection detected: value starts with '-' or '+'")
}

func TestCheckForArgumentInjection_MinusNumber(t *testing.T) {
	// tail -10
	err := checkForArgumentInjection("-10")
	assert.NoError(t, err, "Should allow negative numbers")
}
