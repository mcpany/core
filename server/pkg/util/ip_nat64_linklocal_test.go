package util

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIP_NAT64_LinkLocal(t *testing.T) {
	// 169.254.169.254 (Link Local) embedded in NAT64 prefix (64:ff9b::/96)
	// 169.254.169.254 -> a9.fe.a9.fe
	// 64:ff9b::a9fe:a9fe
	ip := net.ParseIP("64:ff9b::a9fe:a9fe")
	assert.NotNil(t, ip)

	// IsPrivateIP should return true for NAT64 link-local
	assert.True(t, IsPrivateIP(ip), "IsPrivateIP should return true for NAT64 link-local")
}

func TestIsPrivateIP_NAT64_Loopback(t *testing.T) {
	// 127.0.0.1 embedded in NAT64
	// 127.0.0.1 -> 7f.00.00.01
	// 64:ff9b::7f00:01
	ip := net.ParseIP("64:ff9b::7f00:01")
	assert.NotNil(t, ip)

	assert.True(t, IsPrivateIP(ip), "IsPrivateIP should return true for NAT64 loopback")
}

func TestIsPrivateNetworkIP_NAT64_LinkLocal_Behavior(t *testing.T) {
	// Documenting that IsPrivateNetworkIP currently excludes NAT64 LinkLocal,
    // consistent with "It does NOT include loopback or link-local addresses."
	ip := net.ParseIP("64:ff9b::a9fe:a9fe")
	assert.False(t, IsPrivateNetworkIP(ip), "IsPrivateNetworkIP should return false for link-local (even if NAT64)")
}
