// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestExtractIP_ZoneIndex(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		shouldBePriv  bool
		shouldBeValid bool // can be parsed by net.ParseIP
	}{
		{
			name:          "IPv6 Link Local with Zone Index",
			input:         "fe80::1%eth0",
			expected:      "fe80::1",
			shouldBePriv:  true,
			shouldBeValid: true,
		},
		{
			name:          "IPv6 Link Local with Zone Index and Brackets",
			input:         "[fe80::1%eth0]",
			expected:      "fe80::1",
			shouldBePriv:  true,
			shouldBeValid: true,
		},
		{
			name:          "IPv6 Link Local with Zone Index, Brackets and Port",
			input:         "[fe80::1%eth0]:8080",
			expected:      "fe80::1",
			shouldBePriv:  true,
			shouldBeValid: true,
		},
		{
			name:          "IPv6 Link Local with Zone Index, no brackets, and Port (invalid URL but maybe passed)",
			input:         "fe80::1%eth0:8080",
			// SplitHostPort might fail or return weird stuff.
			// "fe80::1%eth0:8080" -> too many colons.
			// ip = input.
			// % found at index 7. -> "fe80::1"
			expected:      "fe80::1",
			shouldBePriv:  true,
			shouldBeValid: true,
		},
		{
			name:          "Normal IPv6",
			input:         "2001:db8::1",
			expected:      "2001:db8::1",
			shouldBePriv:  true, // Documentation prefix is in privateNetworkBlocks
			shouldBeValid: true,
		},
		{
			name:          "Zone index only",
			input:         "%",
			expected:      "",
			shouldBePriv:  false, // ParseIP("") -> nil
			shouldBeValid: false,
		},
		{
			name:          "Address starting with zone index",
			input:         "%eth0",
			expected:      "",
			shouldBePriv:  false,
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIP(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractIP(%q) = %q, want %q", tt.input, got, tt.expected)
			}

			ip := net.ParseIP(got)
			if tt.shouldBeValid && ip == nil {
				t.Errorf("net.ParseIP(%q) returned nil, expected valid IP", got)
			} else if !tt.shouldBeValid && ip != nil {
				t.Errorf("net.ParseIP(%q) returned %v, expected nil", got, ip)
			}

			if tt.shouldBeValid {
				isPriv := IsPrivateIP(ip)
				if isPriv != tt.shouldBePriv {
					t.Errorf("IsPrivateIP(%v) = %v, want %v", ip, isPriv, tt.shouldBePriv)
				}
			}
		})
	}
}
