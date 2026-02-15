package tool

import (
	"reflect"
	"testing"
)

func TestParseURLSegments(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []urlSegment
	}{
		{
			name:     "Empty",
			template: "",
			want:     []urlSegment{},
		},
		{
			name:     "Literal only",
			template: "http://example.com/foo",
			want: []urlSegment{
				{isParam: false, value: "http://example.com/foo"},
			},
		},
		{
			name:     "Single param",
			template: "http://example.com/{{foo}}",
			want: []urlSegment{
				{isParam: false, value: "http://example.com/"},
				{isParam: true, value: "foo"},
			},
		},
		{
			name:     "Mixed params and literals",
			template: "http://example.com/{{foo}}/bar/{{baz}}",
			want: []urlSegment{
				{isParam: false, value: "http://example.com/"},
				{isParam: true, value: "foo"},
				{isParam: false, value: "/bar/"},
				{isParam: true, value: "baz"},
			},
		},
		{
			name:     "Unclosed template (Bug Fix)",
			template: "http://example.com/foo{{bar",
			want: []urlSegment{
				{isParam: false, value: "http://example.com/foo"},
				{isParam: false, value: "{{bar"},
			},
		},
		{
			name:     "Multiple unclosed",
			template: "foo{{bar}}baz{{qux",
			want: []urlSegment{
				{isParam: false, value: "foo"},
				{isParam: true, value: "bar"},
				{isParam: false, value: "baz"},
				{isParam: false, value: "{{qux"},
			},
		},
		{
			name:     "Param with spaces",
			template: "{{ param }}",
			want: []urlSegment{
				{isParam: true, value: " param "},
			},
		},
		{
			name:     "Suffix after param",
			template: "{{param}}suffix",
			want: []urlSegment{
				{isParam: true, value: "param"},
				{isParam: false, value: "suffix"},
			},
		},
		{
			name:     "Broken template",
			template: "{{",
			want: []urlSegment{
				{isParam: false, value: "{{"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseURLSegments(tt.template)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseURLSegments(%q) = %v, want %v", tt.template, got, tt.want)
			}
		})
	}
}
