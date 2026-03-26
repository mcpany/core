// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestRedactSlice_MultipleDirty ...
// Summary: TestRedactSlice_MultipleDirty
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestRedactSlice_NestedSliceDirty ...
// Summary: TestRedactSlice_NestedSliceDirty
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
