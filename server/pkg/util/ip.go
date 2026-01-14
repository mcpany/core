// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	"context"
	"net"
)

type contextKey string

const remoteIPContextKey = contextKey("remote_ip")

// ContextWithRemoteIP returns a new context with the remote IP.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// ExtractIP extracts the IP address from a host:port string or just an IP string.
// It also handles IPv6 brackets and strips IPv6 zone indices (e.g., %eth0).
func ExtractIP(addr string) string {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}
	if len(ip) > 0 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		ip = ip[1 : len(ip)-1]
	}
	// Strip zone index if present (e.g. fe80::1%eth0 -> fe80::1)
	for i := 0; i < len(ip); i++ {
		if ip[i] == '%' {
			return ip[:i]
		}
	}
	return ip
}

// RemoteIPFromContext retrieves the remote IP from the context.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

var (
	privateNetworkBlocks []*net.IPNet
	linkLocalBlocks      []*net.IPNet
)

func init() {
	// RFC1918 + RFC4193 + RFC6598 (CGNAT) + others
	for _, cidr := range []string{
		"0.0.0.0/8",      // Current network (RFC 1122)
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"100.64.0.0/10",  // RFC6598 Shared Address Space (CGNAT)
		"192.0.0.0/24",   // IETF Protocol Assignments (RFC 6890)
		"192.0.2.0/24",   // TEST-NET-1 (RFC 5737)
		"198.18.0.0/15",  // Benchmarking (RFC 2544)
		"198.51.100.0/24", // TEST-NET-2 (RFC 5737)
		"203.0.113.0/24", // TEST-NET-3 (RFC 5737)
		"240.0.0.0/4",    // Class E (RFC 1112)
		"255.255.255.255/32", // Broadcast
		"fc00::/7",       // RFC4193 unique local address
		"2001:db8::/32",  // IPv6 documentation (RFC 3849)
		"64:ff9b::/96",   // IPv4/IPv6 translation (RFC 6052)
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateNetworkBlocks = append(privateNetworkBlocks, block)
		}
	}

	// Link Local
	for _, cidr := range []string{
		"169.254.0.0/16", // RFC3927 link-local
		"fe80::/10",      // RFC4291 link-local
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			linkLocalBlocks = append(linkLocalBlocks, block)
		}
	}
}

// IsPrivateNetworkIP checks if the IP address is a private network address.
// This includes RFC1918, RFC4193 (Unique Local), and RFC6598 (CGNAT).
// It does NOT include loopback or link-local addresses.
func IsPrivateNetworkIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		// IPv4 fast path: check directly to avoid linear scan of net.IPNet slices
		return isPrivateNetworkIPv4(ip4)
	}

	for _, block := range privateNetworkBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// isPrivateNetworkIPv4 checks if an IPv4 address is private.
// ip must be a valid 4-byte IPv4 address slice.
func isPrivateNetworkIPv4(ip net.IP) bool {
	switch ip[0] {
	case 0:
		return true // 0.0.0.0/8
	case 10:
		return true // 10.0.0.0/8
	case 100:
		return ip[1] >= 64 && ip[1] <= 127 // 100.64.0.0/10
	case 172:
		return ip[1] >= 16 && ip[1] <= 31 // 172.16.0.0/12
	case 192:
		if ip[1] == 168 {
			return true // 192.168.0.0/16
		}
		if ip[1] == 0 {
			return ip[2] == 0 || ip[2] == 2 // 192.0.0.0/24 or 192.0.2.0/24
		}
	case 198:
		if ip[1] == 18 || ip[1] == 19 {
			return true // 198.18.0.0/15
		}
		return ip[1] == 51 && ip[2] == 100 // 198.51.100.0/24
	case 203:
		return ip[1] == 0 && ip[2] == 113 // 203.0.113.0/24
	}

	// Class E (240.0.0.0/4) and Broadcast (255.255.255.255)
	if ip[0] >= 240 {
		return true
	}

	return false
}

// IsPrivateIP checks if the IP address is a private, link-local, or loopback address.
func IsPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	if IsPrivateNetworkIP(ip) {
		return true
	}
	for _, block := range linkLocalBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
