// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"reflect"
	"testing"
)

func TestParseQueryManual(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		want     []queryPart
	}{
		{
			name:     "empty",
			rawQuery: "",
			want:     nil,
		},
		{
			name:     "simple",
			rawQuery: "foo=bar",
			want: []queryPart{
				{raw: "foo=bar", key: "foo", keyDecoded: true},
			},
		},
		{
			name:     "multiple",
			rawQuery: "foo=bar&baz=qux",
			want: []queryPart{
				{raw: "foo=bar", key: "foo", keyDecoded: true},
				{raw: "baz=qux", key: "baz", keyDecoded: true},
			},
		},
		{
			name:     "invalid_encoding",
			rawQuery: "foo=%GG",
			want: []queryPart{
				{raw: "foo=%GG", key: "foo", keyDecoded: true, isInvalid: true},
			},
		},
		{
			name:     "key_value_encoded",
			rawQuery: "foo%20bar=baz%20qux",
			want: []queryPart{
				{raw: "foo%20bar=baz%20qux", key: "foo bar", keyDecoded: true},
			},
		},
		{
			name:     "no_value",
			rawQuery: "foo",
			want: []queryPart{
				{raw: "foo", key: "foo", keyDecoded: true},
			},
		},
		{
			name:     "empty_parts",
			rawQuery: "&foo=bar&&",
			want: []queryPart{
				{raw: "foo=bar", key: "foo", keyDecoded: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseQueryManual(tt.rawQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseQueryManual() = %v, want %v", got, tt.want)
			}
		})
	}
}
