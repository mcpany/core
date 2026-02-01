// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stringerStruct struct {
	Val string
}

func (s stringerStruct) String() string {
	return "S:" + s.Val
}

type plainStruct struct {
	ID int
}

func TestTransformer_Reflection(t *testing.T) {
	t.Parallel()
	transformer := NewTransformer()

	tests := []struct {
		name        string
		templateStr string
		data        map[string]any
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "typed slice int",
			templateStr: `{{join "," .items}}`,
			data:        map[string]any{"items": []int{1, 2, 3}},
			want:        "1,2,3",
			wantErr:     false,
		},
		{
			name:        "typed slice string",
			templateStr: `{{join "|" .items}}`,
			data:        map[string]any{"items": []string{"a", "b", "c"}},
			want:        "a|b|c",
			wantErr:     false,
		},
		{
			name:        "array int",
			templateStr: `{{join "-" .items}}`,
			data:        map[string]any{"items": [3]int{7, 8, 9}},
			want:        "7-8-9",
			wantErr:     false,
		},
		{
			name:        "invalid input not slice",
			templateStr: `{{join "," .items}}`,
			data:        map[string]any{"items": 123},
			wantErr:     true,
			errContains: "expected slice or array, got int",
		},
		{
			name:        "stringer interface",
			templateStr: `{{join ";" .items}}`,
			data: map[string]any{
				"items": []stringerStruct{
					{Val: "one"},
					{Val: "two"},
				},
			},
			want:    "S:one;S:two",
			wantErr: false,
		},
		{
			name:        "default formatting (plain struct)",
			templateStr: `{{join " " .items}}`,
			data: map[string]any{
				"items": []plainStruct{
					{ID: 1},
					{ID: 2},
				},
			},
			want:    "{1} {2}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformer.Transform(tt.templateStr, tt.data)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
