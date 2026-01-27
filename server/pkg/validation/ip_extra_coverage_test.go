package validation

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLoopback_Extended(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"127.1.2.3", true},
		{"::ffff:127.0.0.1", true}, // IPv4-mapped
		{"::127.0.0.1", true},      // IPv4-compatible
		{"64:ff9b::127.0.0.1", true}, // NAT64
		{"8.8.8.8", false},
		{"::2", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.Equal(t, tt.want, IsLoopback(ip))
		})
	}
}

func TestIsLinkLocal_Extended(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"169.254.0.1", true},
		{"fe80::1", true},
		{"::ffff:169.254.0.1", true}, // IPv4-mapped
		{"::169.254.0.1", true},      // IPv4-compatible
		{"64:ff9b::169.254.0.1", true}, // NAT64
		{"10.0.0.1", false},
		{"8.8.8.8", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.Equal(t, tt.want, IsLinkLocal(ip))
		})
	}
}

func TestIsPrivateNetworkIP_Extended(t *testing.T) {
	// 0.0.0.0 and :: are unspecified, treated as private
	assert.True(t, IsPrivateNetworkIP(net.ParseIP("0.0.0.0")))
	assert.True(t, IsPrivateNetworkIP(net.ParseIP("::")))
}
