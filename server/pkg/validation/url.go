// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

const trueVal = "true"

var (
	// ErrUnsafeIP is the base error for unsafe IP addresses.
	ErrUnsafeIP    = errors.New("unsafe IP address")
	ErrLoopback    = fmt.Errorf("%w: loopback address is not allowed", ErrUnsafeIP)
	ErrPrivate     = fmt.Errorf("%w: private network address is not allowed", ErrUnsafeIP)
	ErrLinkLocal   = fmt.Errorf("%w: link-local address is not allowed (metadata service protection)", ErrUnsafeIP)
	ErrMulticast   = fmt.Errorf("%w: multicast address is not allowed", ErrUnsafeIP)
	ErrUnspecified = fmt.Errorf("%w: unspecified address (0.0.0.0) is not allowed", ErrUnsafeIP)
)

// IsSafeIP checks if the IP address string is safe to connect to,
// respecting the allowed network resources policy.
//
// IsSafeIP is a variable to allow mocking in tests.
var IsSafeIP = func(ipStr string) error {
	// Bypass if explicitly allowed (for testing/development)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == trueVal {
		return nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address")
	}

	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == trueVal
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == trueVal

	return ValidateIP(ip, allowLoopback, allowPrivate)
}

// LookupIP is a variable to allow mocking DNS resolution in tests.
var LookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, network, host)
}

// ValidateHost checks if the host resolves to a safe IP.
// It handles both IP literals and domain names.
func ValidateHost(host string) error {
	// Bypass if explicitly allowed (for testing/development)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == trueVal {
		return nil
	}

	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == trueVal
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == trueVal

	// Check if host is an IP literal
	if ip := net.ParseIP(host); ip != nil {
		return ValidateIP(ip, allowLoopback, allowPrivate)
	}

	// Resolve Domain
	// Use a timeout for resolution
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", host, err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for host %q", host)
	}

	// Check all resolved IPs
	for _, ip := range ips {
		if err := ValidateIP(ip, allowLoopback, allowPrivate); err != nil {
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
	// Bypass if explicitly allowed (for testing/development)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == trueVal {
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

	return ValidateHost(host)
}

// ValidateIP checks if the IP address is allowed based on the policy.
func ValidateIP(ip net.IP, allowLoopback, allowPrivate bool) error {
	if !allowLoopback && (ip.IsLoopback() || IsNAT64Loopback(ip) || (IsIPv4Compatible(ip) && ip[12] == 127)) {
		return ErrLoopback
	}
	if ip.IsLinkLocalUnicast() || IsNAT64LinkLocal(ip) || (IsIPv4Compatible(ip) && ip[12] == 169 && ip[13] == 254) {
		return ErrLinkLocal
	}
	if ip.IsLinkLocalMulticast() {
		return fmt.Errorf("%w: link-local multicast address is not allowed", ErrUnsafeIP)
	}
	if ip.IsMulticast() {
		return ErrMulticast
	}
	if ip.IsUnspecified() {
		return ErrUnspecified
	}
	if !allowPrivate && IsPrivateNetworkIP(ip) {
		return ErrPrivate
	}
	return nil
}
