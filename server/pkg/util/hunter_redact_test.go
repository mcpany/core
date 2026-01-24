// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"strings"
	"testing"
)

func TestRedactDSN_QueryParam(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Redis with password and query params (no user)",
			input:    "redis://:password?param=value",
			expected: "redis://:[REDACTED]?param=value",
		},
		{
			name:     "Redis with password and query params (with user)",
			input:    "redis://user:password?param=value",
			expected: "redis://user:[REDACTED]?param=value",
		},
		{
			name:     "Redis with password containing question mark (encoded)",
			input:    "redis://:pass%3Fword?param=value",
			expected: "redis://:[REDACTED]?param=value",
		},
		{
			name:     "Redis with literal ? in password (ambiguous)",
			input:    "redis://:pass?word?param=value",
			expected: "redis://:[REDACTED]?word?param=value",
		},
		{
			name:     "Postgres with # in password (invalid port error in url.Parse)",
			input:    "postgres://user:p#ssword@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "Postgres with ? in password (invalid port error in url.Parse)",
			input:    "postgres://user:p?ssword@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "Postgres with / in password (invalid port error in url.Parse)",
			input:    "postgres://user:p/ssword@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactDSN(tt.input)
			if got != tt.expected {
				t.Errorf("RedactDSN(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Trailing dot",
			input:    "file.",
			expected: "file.",
		},
		{
			name:     "Only dots",
			input:    "...",
			expected: "...",
		},
		{
			name:     "Mixed separators",
			input:    "dir\\file",
			expected: "dir_file",
		},
		{
			name:     "Unicode filename",
			input:    "文件名.txt",
			expected: "文件名.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractIP_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv6 literal",
			input:    "::1",
			expected: "::1",
		},
		{
			name:     "IPv6 literal with brackets",
			input:    "[::1]",
			expected: "::1",
		},
		{
			name:     "IPv6 literal with zone",
			input:    "::1%eth0",
			expected: "::1",
		},
		{
			name:     "IPv6 literal with zone and brackets",
			input:    "[::1%eth0]",
			expected: "::1",
		},
		{
			name:     "Garbage with percent",
			input:    "127.0.0.1%25",
			expected: "127.0.0.1",
		},
		{
			name:     "Garbage with percent and brackets",
			input:    "[127.0.0.1%25]", // Brackets are stripped if both present
			expected: "127.0.0.1",
		},
		{
			name:     "Malformed brackets (start only)",
			input:    "[::1",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIP(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractIP(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsPrivateIP_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "IPv4-compatible loopback",
			ip:       "::127.0.0.1",
			expected: true,
		},
		{
			name:     "IPv4-compatible private",
			ip:       "::10.0.0.1",
			expected: true,
		},
		{
			name:     "IPv4-mapped loopback",
			ip:       "::ffff:127.0.0.1",
			expected: true,
		},
		{
			name:     "IPv4-mapped private",
			ip:       "::ffff:192.168.1.1",
			expected: true,
		},
		{
			name:     "NAT64 loopback",
			ip:       "64:ff9b::7f00:0001", // 127.0.0.1
			expected: true,
		},
		{
			name:     "NAT64 private",
			ip:       "64:ff9b::c0a8:0101", // 192.168.1.1
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			if got := IsPrivateIP(ip); got != tt.expected {
				t.Errorf("IsPrivateIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}

func TestLevenshteinDistance_Unicode(t *testing.T) {
	tests := []struct {
		s1, s2 string
		want   int
	}{
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"你好", "你好", 0},
		{"你好", "你", 1},
		{"你好", "好", 1},
		{"kitten", "sitting", 3},
		{strings.Repeat("a", 300), strings.Repeat("a", 300), 0},
	}
	for _, tt := range tests {
		got := LevenshteinDistance(tt.s1, tt.s2)
		if got != tt.want {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
		}
	}
}
