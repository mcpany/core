// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFastMarshalToString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		obj := map[string]interface{}{
			"key": "value",
			"num": 42,
		}
		str, err := FastMarshalToString(obj)
		assert.NoError(t, err)
		assert.Contains(t, str, `"key":"value"`)
		assert.Contains(t, str, `"num":42`)
	})

	t.Run("failure", func(t *testing.T) {
		obj := map[string]interface{}{
			"func": func() {},
		}
		str, err := FastMarshalToString(obj)
		assert.Error(t, err)
		assert.Empty(t, str)
	})
}
