// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CustomInt int

func (c CustomInt) String() string {
	return fmt.Sprintf("Custom(%d)", int(c))
}

func TestTransformer_JoinEdgeCases(t *testing.T) {
	tr := NewTransformer()

	t.Run("nil_in_slice", func(t *testing.T) {
		items := []any{1, nil, 3}
		data := map[string]any{"items": items}
		res, err := tr.Transform(`{{join "," .items}}`, data)
		assert.NoError(t, err)
		assert.Equal(t, "1,<nil>,3", string(res))
	})

	t.Run("custom_stringer", func(t *testing.T) {
		items := []any{CustomInt(1), CustomInt(2)}
		data := map[string]any{"items": items}
		res, err := tr.Transform(`{{join "," .items}}`, data)
		assert.NoError(t, err)
		assert.Equal(t, "Custom(1),Custom(2)", string(res))
	})

	t.Run("time_duration", func(t *testing.T) {
		items := []any{1 * time.Second, 500 * time.Millisecond}
		data := map[string]any{"items": items}
		res, err := tr.Transform(`{{join "," .items}}`, data)
		assert.NoError(t, err)
		assert.Equal(t, "1s,500ms", string(res))
	})

	t.Run("mixed_types_with_nil", func(t *testing.T) {
		items := []any{"text", nil, 42, 10 * time.Minute}
		data := map[string]any{"items": items}
		res, err := tr.Transform(`{{join "|" .items}}`, data)
		assert.NoError(t, err)
		assert.Equal(t, "text|<nil>|42|10m0s", string(res))
	})
}

func TestTransformer(t *testing.T) {
	t.Parallel()
	transformer := NewTransformer()

	tests := []struct {
		name        string
		templateStr string
		data        map[string]any
		want        string
		wantErr     bool
	}{
		{
			name:        "simple template",
			templateStr: "Hello, {{.name}}!",
			data:        map[string]any{"name": "world"},
			want:        "Hello, world!",
			wantErr:     false,
		},
		{
			name:        "json template",
			templateStr: `{"name": "{{.name}}", "age": {{.age}}}`,
			data:        map[string]any{"name": "John", "age": 30},
			want:        `{"name": "John", "age": 30}`,
			wantErr:     false,
		},
		{
			name:        "with join function",
			templateStr: `{{join "," .items}}`,
			data:        map[string]any{"items": []any{"a", "b", "c"}},
			want:        "a,b,c",
			wantErr:     false,
		},
		{
			name:        "with join function mixed types",
			templateStr: `{{join "," .items}}`,
			data: map[string]any{
				"items": []any{
					"text",
					int(42),
					int8(8),
					int16(16),
					int32(32),
					int64(64),
					uint(42),
					uint8(8),
					uint16(16),
					uint32(32),
					uint64(64),
					float32(3.14),
					float64(3.14159),
					true,
					false,
				},
			},
			want:    "text,42,8,16,32,64,42,8,16,32,64,3.14,3.14159,true,false",
			wantErr: false,
		},
		{
			name:        "invalid template",
			templateStr: `{{.name`,
			data:        map[string]any{"name": "world"},
			want:        "",
			wantErr:     true,
		},
		{
			name:        "missing key",
			templateStr: `Hello, {{.name}}!`,
			data:        map[string]any{"missing": "world"},
			want:        "Hello, <no value>!",
			wantErr:     false,
		},
		{
			name:        "json output",
			templateStr: `{"user": {{json .user}}}`,
			data:        map[string]any{"user": map[string]any{"name": "John", "age": 30}},
			want:        `{"user": {"age":30,"name":"John"}}`,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformer.Transform(tt.templateStr, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.name == "json template" || tt.name == "json output" {
				assert.JSONEq(t, tt.want, string(got))
			} else {
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}
