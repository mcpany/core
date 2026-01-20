// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIP_NAT64_Loopback(t *testing.T) {
	// 64:ff9b::127.0.0.1
	// 127.0.0.1 = 7f 00 00 01
	// So 64:ff9b::7f00:1
	ip := net.ParseIP("64:ff9b::127.0.0.1")
	assert.NotNil(t, ip)

	// This currently fails because IsPrivateIP returns false for this IP
	assert.True(t, IsPrivateIP(ip), "NAT64 wrapped loopback should be considered private")
}

func TestIsPrivateIP_NAT64_LinkLocal(t *testing.T) {
	// 64:ff9b::169.254.1.1
	ip := net.ParseIP("64:ff9b::169.254.1.1")
	assert.NotNil(t, ip)

	// This currently fails because IsPrivateIP returns false for this IP
	assert.True(t, IsPrivateIP(ip), "NAT64 wrapped link-local should be considered private")
}
