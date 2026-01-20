// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestIsPrivateNetworkIP_Multicast(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
		reason   string
	}{
		// IPv4 Multicast
		{"224.0.0.1", false, "Link Local (224.0.0.x) - excluded by IsPrivateNetworkIP"},
		{"239.1.1.1", true, "Admin Scoped (239.0.0.0/8) - Private"},
		{"239.255.255.255", true, "Admin Scoped (239.0.0.0/8) - Private"},
		{"224.0.1.1", false, "Global Multicast (Internetwork Control) - Public"},
		{"232.0.0.1", false, "Source Specific Multicast - Public"},

		// IPv6 Multicast
		// ff0s::
		// Scopes: 0(R), 1(I), 2(L), 3(Realm), 4(Admin), 5(Site), 8(Org), E(Global), F(R)

		{"ff01::1", false, "Interface Local (Scope 1) - Treated as Loopback (excluded)"},
		{"ff02::1", false, "Link Local (Scope 2) - Excluded by IsPrivateNetworkIP"},
		{"ff03::1", true, "Realm Local (Scope 3) - Private"},
		{"ff04::1", true, "Admin Local (Scope 4) - Private"},
		{"ff05::1", true, "Site Local (Scope 5) - Private"},
		{"ff08::1", true, "Org Local (Scope 8) - Private"},
		{"ff0e::1", false, "Global (Scope E) - Public"},
		{"ff00::1", true, "Reserved (Scope 0) - Treated as Private"},
		{"ff0f::1", true, "Reserved (Scope F) - Treated as Private"},

		// IPv4-mapped IPv6
		{"::ffff:239.1.1.1", true, "IPv4-mapped Admin Scoped - Private"},
		{"::ffff:224.0.1.1", false, "IPv4-mapped Global - Public"},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.Equal(t, tt.expected, IsPrivateNetworkIP(ip), tt.reason)
		})
	}
}
