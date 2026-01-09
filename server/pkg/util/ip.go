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
// It also handles IPv6 brackets.
func ExtractIP(addr string) string {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}
	if len(ip) > 0 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		ip = ip[1 : len(ip)-1]
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
	for _, block := range privateNetworkBlocks {
		if block.Contains(ip) {
			return true
		}
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
