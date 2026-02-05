// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTokensInValue_MapCycle(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	// Create a map that contains itself
	m := make(map[string]interface{})
	m["self"] = m

	_, err := CountTokensInValue(tokenizer, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestCountTokensInValue_SliceCycle(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	// Create a slice that contains itself
	s := make([]interface{}, 1)
	s[0] = s

	_, err := CountTokensInValue(tokenizer, s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}
