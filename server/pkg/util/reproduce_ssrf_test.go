// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestIPv4MappedLoopbackBypass(t *testing.T) {
	// ::ffff:127.0.0.1 is an IPv4-mapped IPv6 address pointing to localhost
	ipStr := "::ffff:127.0.0.1"
	ip := net.ParseIP(ipStr)
	assert.NotNil(t, ip)

	// 1. Check if Go considers it a loopback address directly
	isLoopback := ip.IsLoopback()
	t.Logf("IsLoopback(%s) = %v", ipStr, isLoopback)

	// 2. Check if IsPrivateNetworkIP considers it private
	// It should return FALSE because loopback is not considered "Private Network" (RFC1918)
	// But it is dangerous for SSRF.
	isPrivate := validation.IsPrivateNetworkIP(ip)
	t.Logf("IsPrivateNetworkIP(%s) = %v", ipStr, isPrivate)

	// 3. Simulate SafeDialer check
	// SafeDialer logic:
	// if !d.AllowLoopback && (ip.IsLoopback() || isNAT64Loopback(ip) || ip.IsUnspecified()) { block }
	// if !d.AllowLinkLocal && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || isNAT64LinkLocal(ip)) { block }
	// if !d.AllowPrivate && IsPrivateNetworkIP(ip) { block }

	dialer := NewSafeDialer()
	// We want to block loopback
	dialer.AllowLoopback = false
	dialer.AllowPrivate = false

	// Manually run the check logic from SafeDialer.DialContext
	blocked := false
	if !dialer.AllowLoopback && (ip.IsLoopback() || isNAT64Loopback(ip) || ip.IsUnspecified()) {
		blocked = true
		t.Log("Blocked by Loopback check")
	} else if !dialer.AllowPrivate && validation.IsPrivateNetworkIP(ip) {
		blocked = true
		t.Log("Blocked by Private Network check")
	}

	if !blocked {
		t.Errorf("VULNERABILITY CONFIRMED: %s bypassed SafeDialer checks!", ipStr)
	} else {
		t.Log("Safe: Address was blocked.")
	}
}
