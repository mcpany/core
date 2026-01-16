package util

import (
	"net"
	"testing"
)

func TestIsPrivateIP_NAT64_Fix(t *testing.T) {
	// 64:ff9b::192.168.1.1 is a NAT64 address mapping to a private IPv4 address.
	// It should be considered private by IsPrivateIP.
	ipStr := "64:ff9b::192.168.1.1"
	ip := net.ParseIP(ipStr)
	if ip == nil {
		t.Fatalf("Failed to parse IP: %s", ipStr)
	}

	if !IsPrivateIP(ip) {
		t.Errorf("IsPrivateIP(%s) = false, want true (NAT64 mapped private IP should be private)", ipStr)
	}
}
