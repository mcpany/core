package transformer

import (
	"testing"
)

func TestJoinFunc_Detailed(t *testing.T) {
	tests := []struct {
		name     string
		sep      string
		input    any
		expected string
		wantErr  bool
	}{
		{
			name:     "string slice",
			sep:      ",",
			input:    []string{"a", "b", "c"},
			expected: "a,b,c",
		},
		{
			name:     "int slice",
			sep:      ",",
			input:    []int{1, 2, 3},
			expected: "1,2,3",
		},
		{
			name:     "mixed any slice",
			sep:      "|",
			input:    []any{"a", 1, true, 3.14},
			expected: "a|1|true|3.14",
		},
		{
			name:     "nil in any slice",
			sep:      ",",
			input:    []any{"a", nil, "b"},
			expected: "a,<nil>,b",
		},
		{
			name:     "typed nil in any slice",
			sep:      ",",
			input:    []any{"a", (*int)(nil), "b"},
			expected: "a,<nil>,b",
		},
		{
			name:     "empty slice",
			sep:      ",",
			input:    []string{},
			expected: "",
		},
		{
			name:     "single element",
			sep:      ",",
			input:    []string{"foo"},
			expected: "foo",
		},
		{
			name:     "pointer slice",
			sep:      ",",
			input:    []*int{},
			expected: "",
		},
		{
			name:    "not a slice",
			sep:     ",",
			input:   "not a slice",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := joinFunc(tt.sep, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("joinFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("joinFunc() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestJoinFunc_Stringer(t *testing.T) {
	input := []any{stringer{val: "foo"}, stringer{val: "bar"}}
	got, err := joinFunc(",", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "foo,bar"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

type stringer struct {
	val string
}

func (s stringer) String() string {
	return s.val
}
