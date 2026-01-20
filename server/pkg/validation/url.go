// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

// IsSafeURL checks if the URL is safe to connect to.
// It validates the scheme and resolves the host to ensure it doesn't point to
// loopback, link-local, private, or multicast addresses.
//
// NOTE: This function performs DNS resolution to check the IP.
// It is susceptible to DNS rebinding attacks if the check is separated from the connection.
// For critical security, use a custom Dialer that validates the IP after resolution.
//
// IsSafeURL is a variable to allow mocking in tests.
var IsSafeURL = func(urlStr string) error {
	// Bypass if explicitly allowed (for testing/development)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == "true" {
		return nil
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// 1. Check Scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s (only http and https are allowed)", u.Scheme)
	}

	// 2. Resolve Host
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL is missing host")
	}

	// Check if host is an IP literal
	if ip := net.ParseIP(host); ip != nil {
		return validateIP(ip)
	}

	// Resolve Domain
	// Use a timeout for resolution
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", host, err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for host %q", host)
	}

	// Check all resolved IPs
	for _, ip := range ips {
		if err := validateIP(ip); err != nil {
			return fmt.Errorf("host %q resolves to unsafe IP %s: %w", host, ip.String(), err)
		}
	}

	return nil
}

func validateIP(ip net.IP) error {
	if isPrivateIP(ip) {
		return fmt.Errorf("private, loopback, or link-local address is not allowed")
	}
	if ip.IsMulticast() {
		return fmt.Errorf("multicast address is not allowed")
	}
	return nil
}

// Helpers copied from server/pkg/util/ip.go to avoid circular dependency.

var privateNetworkBlocksIPv6 []*net.IPNet

func init() {
	// RFC4193 + others
	// Note: IPv4 blocks are handled by isPrivateNetworkIPv4 fast path, so we only need IPv6 here.
	for _, cidr := range []string{
		"fc00::/7",      // RFC4193 unique local address
		"2001:db8::/32", // IPv6 documentation (RFC 3849)
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateNetworkBlocksIPv6 = append(privateNetworkBlocksIPv6, block)
		}
	}
}

// isPrivateIP checks if the IP address is a private, link-local, or loopback address.
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsUnspecified() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		// Link-local (169.254.0.0/16)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		return isPrivateNetworkIPv4(ip4)
	}

	// IPv6 Link-local (fe80::/10)
	if len(ip) == net.IPv6len && ip[0] == 0xfe && ip[1]&0xc0 == 0x80 {
		return true
	}

	// Check for IPv4-compatible IPv6 addresses (::a.b.c.d) for Loopback/Link-local
	if isIPv4Compatible(ip) {
		ip4 := ip[12:16]
		// Loopback (127.0.0.0/8)
		if ip4[0] == 127 {
			return true
		}
		// Link-local (169.254.0.0/16)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}

	return isPrivateNetworkIP(ip)
}

// IsPrivateNetworkIP checks if the IP address is a private network address.
// This includes RFC1918, RFC4193 (Unique Local), and RFC6598 (CGNAT).
// It does NOT include loopback or link-local addresses.
func isPrivateNetworkIP(ip net.IP) bool {
	// Treat unspecified addresses (0.0.0.0 and ::) as private.
	// 0.0.0.0 is also covered by isPrivateNetworkIPv4, but :: wasn't.
	if ip.IsUnspecified() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		// IPv4 fast path: check directly to avoid linear scan of net.IPNet slices
		return isPrivateNetworkIPv4(ip4)
	}

	if isNAT64(ip) || isIPv4Compatible(ip) {
		ip4 := ip[12:16]
		return isPrivateNetworkIPv4(ip4)
	}

	for _, block := range privateNetworkBlocksIPv6 {
		if block.Contains(ip) {
			return true
		}
	}

	// IPv6 Multicast (ff00::/8)
	// We check the scope field (last 4 bits of the second byte).
	// Global scope is 0xE. Link-Local scope is 0x2.
	// We consider all non-global and non-link-local multicast addresses as private.
	// We also exclude Scope 1 (Interface-Local) as it is treated as loopback (which is excluded by contract).
	// Scope 0 (Reserved) and F (Reserved) are treated as Private.
	if len(ip) == net.IPv6len && ip[0] == 0xff {
		scope := ip[1] & 0x0f
		// Exclude Global (0xE), Link-Local (0x2), and Interface-Local (0x1)
		if scope != 0x0e && scope != 0x02 && scope != 0x01 {
			return true
		}
	}

	return false
}

func isNAT64(ip net.IP) bool {
	// Check for NAT64 (IPv4-embedded IPv6) - 64:ff9b::/96 (RFC 6052)
	// If it matches, we extract the last 4 bytes and check if they are private.
	// 64:ff9b:: expands to 0064:ff9b:0000:0000:0000:0000 (96 bits)
	return len(ip) == net.IPv6len &&
		ip[0] == 0x00 && ip[1] == 0x64 && ip[2] == 0xff && ip[3] == 0x9b &&
		ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
		ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0
}

func isIPv4Compatible(ip net.IP) bool {
	// Check for IPv4-compatible IPv6 addresses (::a.b.c.d)
	// These are deprecated but can be used to bypass filters if the underlying OS supports them.
	// First 12 bytes are 0.
	return len(ip) == net.IPv6len &&
		ip[0] == 0 && ip[1] == 0 && ip[2] == 0 && ip[3] == 0 &&
		ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
		ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0
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
	case 239:
		return true // 239.0.0.0/8 (Admin Scoped Multicast)
	}

	// Class E (240.0.0.0/4) and Broadcast (255.255.255.255)
	if ip[0] >= 240 {
		return true
	}

	return false
}
