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
	if ip.IsLoopback() {
		return fmt.Errorf("loopback address is not allowed")
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
	if ip.IsUnspecified() {
		return fmt.Errorf("unspecified address (0.0.0.0) is not allowed")
	}
	if ip.IsPrivate() {
		return fmt.Errorf("private address is not allowed (RFC 1918/4193)")
	}
	return nil
}
