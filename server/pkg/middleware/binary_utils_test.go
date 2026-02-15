package middleware

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat32ToBytes_And_Back(t *testing.T) {
	tests := []struct {
		name   string
		input  []float32
	}{
		{
			name:  "Empty",
			input: []float32{},
		},
		{
			name:  "Single",
			input: []float32{1.23},
		},
		{
			name:  "Multiple",
			input: []float32{1.23, 4.56, 7.89},
		},
		{
			name:  "Special Values",
			input: []float32{0.0, -0.0, float32(math.Inf(1)), float32(math.Inf(-1))},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes := float32ToBytes(tc.input)
			assert.Len(t, bytes, len(tc.input)*4)

			output := bytesToFloat32(bytes)
			assert.Equal(t, tc.input, output)
		})
	}
}

func TestBytesToFloat32_InvalidLength(t *testing.T) {
	input := []byte{0x00, 0x00, 0x00} // 3 bytes
	output := bytesToFloat32(input)
	assert.Nil(t, output)
}

func TestFloat32ToBytes_NaN(t *testing.T) {
	// NaN != NaN so we test separately
	input := []float32{float32(math.NaN())}
	bytes := float32ToBytes(input)
	output := bytesToFloat32(bytes)
	assert.Len(t, output, 1)
	assert.True(t, math.IsNaN(float64(output[0])))
}
