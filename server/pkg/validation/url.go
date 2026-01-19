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
// loopback, link-local, or multicast addresses.
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
	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == "true"
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == "true"

	if !allowLoopback {
		if ip.IsLoopback() {
			return fmt.Errorf("loopback address is not allowed")
		}
		if ip.IsUnspecified() {
			return fmt.Errorf("unspecified address (0.0.0.0) is not allowed")
		}
	}

	if ip.IsLinkLocalUnicast() {
		return fmt.Errorf("link-local address is not allowed (metadata service protection)")
	}
	if ip.IsLinkLocalMulticast() {
		return fmt.Errorf("link-local multicast address is not allowed")
	}
	if ip.IsMulticast() {
		return fmt.Errorf("multicast address is not allowed")
	}

	if !allowPrivate {
		if isPrivateNetworkIP(ip) {
			return fmt.Errorf("private network address is not allowed")
		}
	}

	return nil
}

// Helpers copied from util/ip.go to avoid circular dependency

var privateNetworkBlocksIPv6 []*net.IPNet

func init() {
	// RFC4193 + others
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

// isPrivateNetworkIP checks if the IP address is a private network address.
// This includes RFC1918, RFC4193 (Unique Local), and RFC6598 (CGNAT).
// It does NOT include loopback or link-local addresses.
func isPrivateNetworkIP(ip net.IP) bool {
	// Treat unspecified addresses (0.0.0.0 and ::) as private.
	// 0.0.0.0 is also covered by isPrivateNetworkIPv4, but :: wasn't.
	if ip.IsUnspecified() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		// IPv4 fast path
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
	return false
}

func isNAT64(ip net.IP) bool {
	return len(ip) == net.IPv6len &&
		ip[0] == 0x00 && ip[1] == 0x64 && ip[2] == 0xff && ip[3] == 0x9b &&
		ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
		ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0
}

func isIPv4Compatible(ip net.IP) bool {
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
	}

	// Class E (240.0.0.0/4) and Broadcast (255.255.255.255)
	if ip[0] >= 240 {
		return true
	}

	return false
}
