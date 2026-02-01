// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "testing"

func BenchmarkExtractIP(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{"IPv4_NoPort", "192.168.1.1"},
		{"IPv6_NoPort", "2001:db8::1"},
		{"IPv6_NoPort_Brackets", "[2001:db8::1]"},
		{"IPv4_Port", "192.168.1.1:8080"},
		{"IPv6_Port", "[2001:db8::1]:8080"},
		{"Host_NoPort", "localhost"},
		{"Host_Port", "localhost:8080"},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ExtractIP(c.input)
			}
		})
	}
}
