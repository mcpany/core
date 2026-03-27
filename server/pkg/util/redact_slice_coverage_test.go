// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactSlice_MultipleDirty(t *testing.T) {
	// Slice with multiple items needing redaction to trigger "newSlice != nil" path
	input := []interface{}{
		map[string]interface{}{"password": "s1"},
		map[string]interface{}{"password": "s2"},
	}

	redactedMap := RedactMap(map[string]interface{}{"l": input})
	list := redactedMap["l"].([]interface{})

	assert.Equal(t, "[REDACTED]", list[0].(map[string]interface{})["password"])
	assert.Equal(t, "[REDACTED]", list[1].(map[string]interface{})["password"])
}

func TestRedactSlice_NestedSliceDirty(t *testing.T) {
	// Slice containing nested slices that are dirty
	input := []interface{}{
		[]interface{}{map[string]interface{}{"password": "s1"}},
		[]interface{}{map[string]interface{}{"password": "s2"}},
	}

	redactedMap := RedactMap(map[string]interface{}{"l": input})
	list := redactedMap["l"].([]interface{})

	assert.Equal(t, "[REDACTED]", list[0].([]interface{})[0].(map[string]interface{})["password"])
	assert.Equal(t, "[REDACTED]", list[1].([]interface{})[0].(map[string]interface{})["password"])
}
