// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanPathPreserveDoubleSlash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{name: "empty", input: "", expected: "."},
		{name: "root", input: "/", expected: "/"},
		{name: "simple", input: "/foo/bar", expected: "/foo/bar"},
		{name: "trailing slash removal", input: "/foo/bar/", expected: "/foo/bar"},

		// Double slashes
		{name: "double slash", input: "/foo//bar", expected: "/foo//bar"},
		{name: "double slash at root", input: "//foo", expected: "//foo"},
		{name: "triple slash", input: "///foo", expected: "///foo"},
		{name: "trailing double slash", input: "/foo//", expected: "/foo/"}, // Trailing slash is removed, so /foo// -> /foo/

		// Dot segments
		{name: "single dot", input: "/foo/./bar", expected: "/foo/bar"},
		{name: "double dot parent", input: "/foo/../bar", expected: "/bar"},
		{name: "double dot root", input: "/../bar", expected: "/bar"},
		{name: "double dot root 2", input: "/../../bar", expected: "/bar"},

		// Complex cases combining double slashes and dots
		{name: "double slash and dot", input: "/foo//./bar", expected: "/foo//bar"},
		{name: "double slash and parent", input: "/foo//../bar", expected: "/foo/bar"}, // .. pops empty segment
		{name: "parent pops double slash", input: "/foo/bar/..//baz", expected: "/foo//baz"},

		// Relative paths
		{name: "relative simple", input: "foo/bar", expected: "foo/bar"},
		{name: "relative double slash", input: "foo//bar", expected: "foo//bar"},
		{name: "relative dot", input: "foo/./bar", expected: "foo/bar"},
		{name: "relative parent", input: "foo/../bar", expected: "bar"},
		{name: "relative parent start", input: "../foo", expected: "../foo"},
		{name: "relative parent start 2", input: "../../foo", expected: "../../foo"},
		{name: "relative parent after double slash", input: "foo//../bar", expected: "foo/bar"},

		// Edge cases from logic analysis
		{name: "root ..", input: "/..", expected: "/"},
		{name: "root .. ..", input: "/../..", expected: "/"},
		{name: "root .", input: "/.", expected: "/"},
		{name: "relative .", input: ".", expected: "."},
		{name: "relative ..", input: "..", expected: ".."},

		// Trailing slash behavior
		{name: "root trailing", input: "/", expected: "/"},
		{name: "path trailing", input: "/a/", expected: "/a"},
		{name: "path trailing ..", input: "/a/..", expected: "/"},
		{name: "path trailing .. with slash", input: "/a/../", expected: "/"},

		// weird ones
		{name: "weird 1", input: "/a//../b", expected: "/a/b"},
		{name: "weird 2", input: "/a/..//b", expected: "//b"},
		{name: "weird 3", input: "//../a", expected: "//a"},
		{name: "double slash with multiple parents", input: "//../../foo", expected: "//foo"},
		{name: "double slash with parent and double slash", input: "//..//foo", expected: "///foo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanPathPreserveDoubleSlash(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
