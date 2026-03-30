// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"testing"
)

func TestToString_OtherTypes_Extended(t *testing.T) {
	strVar := "pointer_str"
	intVar := 456
	var nilPtr *int = nil

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"Int16", int16(32000), "32000"},
		{"Int32", int32(2000000), "2000000"},
		{"Uint8", uint8(255), "255"},
		{"Uint16", uint16(65000), "65000"},
		{"Uint32", uint32(4000000000), "4000000000"},
		{"Uint64", uint64(1234567890123456), "1234567890123456"},
		{"Error", fmt.Errorf("an error occurred"), "an error occurred"},
		{"Pointer to string", &strVar, "pointer_str"},
		{"Pointer to int", &intVar, "456"},
		{"Nil pointer", nilPtr, "<nil>"},
		{"Slice", []int{1, 2, 3}, "[1 2 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
