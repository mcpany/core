package template

import (
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "simple template",
			tmpl: "Hello, {{.Name}}!",
			args: map[string]any{"Name": "World"},
			want: "Hello, World!",
		},
		{
			name:    "no args",
			tmpl:    "Hello, World!",
			args:    nil,
			want:    "Hello, World!",
		},
		{
			name: "missing arg",
			tmpl: "Hello, {{.Name}}!",
			args: map[string]any{},
			want: "Hello, <no value>!",
		},
		{
			name:    "invalid template",
			tmpl:    "Hello, {{.Name!",
			args:    map[string]any{"Name": "World"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.tmpl, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %v, want %v", got, tt.want)
			}
		})
	}
}
