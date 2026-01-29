package transformer

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformer_Coverage(t *testing.T) {
	t.Parallel()
	transformer := NewTransformer()

	// Test cases to cover intLen and uintLen implementation in transformer.go
	// intLen calls uintLen.
	// uintLen has optimization for <10, <100, <1000, <10000 and a loop for larger numbers.

	largeInts := []any{
		int(1),                 // 1 digit (<10)
		int(12),                // 2 digits (<100)
		int(123),               // 3 digits (<1000)
		int(1234),              // 4 digits (<10000)
		int(12345),             // 5 digits (loop once)
		int(123456),            // 6 digits
		int(1234567),           // 7 digits
		int(12345678),          // 8 digits
		int(123456789),         // 9 digits (loop twice)
		int(1234567890),        // 10 digits
		int64(math.MaxInt64),   // Max int64 (19 digits)
		int64(math.MinInt64),   // Min int64 (negative, 20 chars "-9223...")
		int(-1),                // Negative 1 digit
		int(-12),               // Negative 2 digits
		int(-12345),            // Negative 5 digits
		true,                   // bool true (4 chars)
		false,                  // bool false (5 chars)
	}

	var expectedParts []string
	for _, v := range largeInts {
		expectedParts = append(expectedParts, fmt.Sprintf("%v", v))
	}
	expected := strings.Join(expectedParts, ",")

	data := map[string]any{
		"items": largeInts,
	}

	// This uses joinFunc which calls estimateLen
	templateStr := `{{join "," .items}}`

	got, err := transformer.Transform(templateStr, data)
	require.NoError(t, err)
	assert.Equal(t, expected, string(got))
}

func TestTextParser_ParseJQ_Coverage(t *testing.T) {
	t.Parallel()
	parser := NewTextParser()
	jsonInput := []byte(`{"a": 1, "b": "string"}`)

	t.Run("runtime_error", func(t *testing.T) {
		// Query that causes runtime error: adding number and string
		query := `.a + .b`

		_, err := parser.Parse("jq", jsonInput, nil, query)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "jq execution failed")
	})

	t.Run("empty_result", func(t *testing.T) {
		// Query that returns empty
		query := `select(.a > 5)`

		result, err := parser.Parse("jq", jsonInput, nil, query)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
