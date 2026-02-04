// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterpreterInjection(t *testing.T) {
	// Ruby Backtick Injection
	err := checkForShellInjection("#{system('calc')}", "ruby -e puts `echo {{msg}}`", "{{msg}}", "ruby")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ruby interpolation injection detected")

	// Perl Variable Injection
	err = checkForShellInjection("$ENV{SECRET}", "perl -e print `echo {{msg}}`", "{{msg}}", "perl")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable interpolation injection detected")

	// Perl Array Injection
	err = checkForShellInjection("@INC", "perl -e print `echo {{msg}}`", "{{msg}}", "perl")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable interpolation injection detected")

	// PHP Variable Injection
	err = checkForShellInjection("$secret", "php -r echo `echo {{msg}}`;", "{{msg}}", "php")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable interpolation injection detected")

	// Safe Input
	err = checkForShellInjection("safe_input", "ruby -e puts `echo {{msg}}`", "{{msg}}", "ruby")
	assert.NoError(t, err)
}
