// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const trueVal = "true"

// isTrue checks if the given string represents a truthy value.
// It accepts "true", "TRUE", "True", and "1".
func isTrue(s string) bool {
	return strings.EqualFold(s, "true") || s == "1"
}

// IsSafeURL checks if the URL is safe to connect to.
// It validates the scheme and resolves the host to ensure it doesn't point to
// loopback, link-local, private, or multicast addresses.
//
// NOTE: This function performs DNS resolution to check the IP.
// It is susceptible to DNS rebinding attacks if the check is separated from the connection.
// For critical security, use a custom Dialer that validates the IP after resolution.
//
// lookupIPFunc is a variable to allow mocking DNS resolution in tests.
var lookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, network, host)
}

// IsSafeTarget checks if the URL target (host/IP) is safe to connect to.
// It validates the host/IP against loopback/private/link-local restrictions.
// It does NOT restrict the scheme.
//
// IsSafeTarget is a variable to allow mocking in tests.
var IsSafeTarget = func(urlStr string) error {
	// Bypass if explicitly allowed (for testing/development)
	if isTrue(os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")) {
		return nil
	}

	allowLoopback := isTrue(os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES"))
	allowPrivate := isTrue(os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES"))

	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Resolve Host
	host := u.Hostname()
	if host == "" {
		// If the URL has no host (e.g. mailto:foo@bar), skip IP check?
		// For SSRF protection, we usually care about schemes with hosts.
		// If scheme is generic, maybe skip?
		// However, returning nil here might bypass checks for malformed URLs.
		// Let's assume for now that if we are checking target safety, we expect a host.
		return fmt.Errorf("URL is missing host")
	}

	// Check if host is an IP literal
	if ip := net.ParseIP(host); ip != nil {
		return validateIP(ip, allowLoopback, allowPrivate)
	}

	// Resolve Domain
	// Use a timeout for resolution
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := lookupIPFunc(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", host, err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for host %q", host)
	}

	// Check all resolved IPs
	for _, ip := range ips {
		if err := validateIP(ip, allowLoopback, allowPrivate); err != nil {
			return fmt.Errorf("host %q resolves to unsafe IP %s: %w", host, ip.String(), err)
		}
	}

	return nil
}

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
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// 1. Check Scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s (only http and https are allowed)", u.Scheme)
	}

	// 2. Check Target Safety (Host/IP)
	return IsSafeTarget(urlStr)
}

func validateIP(ip net.IP, allowLoopback, allowPrivate bool) error {
	if !allowLoopback && (ip.IsLoopback() || IsNAT64Loopback(ip) || (IsIPv4Compatible(ip) && ip[12] == 127)) {
		return fmt.Errorf("loopback address is not allowed")
	}
	if ip.IsLinkLocalUnicast() || IsNAT64LinkLocal(ip) || (IsIPv4Compatible(ip) && ip[12] == 169 && ip[13] == 254) {
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
	if !allowPrivate && IsPrivateNetworkIP(ip) {
		return fmt.Errorf("private network address is not allowed")
	}
	return nil
}
