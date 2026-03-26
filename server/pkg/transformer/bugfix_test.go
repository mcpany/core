// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransformer_ReproStringSlice ...
// Summary: TestTransformer_ReproStringSlice
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	transformer := NewTransformer()

	templateStr := `{{join "," .items}}`
	data := map[string]any{
		"items": []string{"a", "b", "c"},
	}

	got, err := transformer.Transform(templateStr, data)
	require.NoError(t, err)
	assert.Equal(t, "a,b,c", string(got))
}

// TestTransformer_ReproIntSlice ...
// Summary: TestTransformer_ReproIntSlice
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	transformer := NewTransformer()

	templateStr := `{{join "," .items}}`
	data := map[string]any{
		"items": []int{1, 2, 3},
	}

	got, err := transformer.Transform(templateStr, data)
	require.NoError(t, err)
	assert.Equal(t, "1,2,3", string(got))
}

// TestTransformer_JoinError ...
// Summary: TestTransformer_JoinError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	transformer := NewTransformer()

	templateStr := `{{join "," .items}}`
	data := map[string]any{
		"items": "not-a-slice",
	}

	got, err := transformer.Transform(templateStr, data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "join: expected slice or array")
	assert.Nil(t, got)
}

// TestTransformer_JsonError ...
// Summary: TestTransformer_JsonError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	transformer := NewTransformer()

	templateStr := `{{json .channel}}`

	// Channels cannot be marshaled to JSON
	data := map[string]any{
		"channel": make(chan int),
	}

	got, err := transformer.Transform(templateStr, data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute template")
	assert.Contains(t, err.Error(), "json: unsupported type")
	assert.Nil(t, got)
}
