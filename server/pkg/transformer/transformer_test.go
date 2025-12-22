package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformer(t *testing.T) {
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
