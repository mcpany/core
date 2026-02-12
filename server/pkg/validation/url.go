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
	return validateURL(urlStr, false)
}

// IsSafeCommandURL is similar to IsSafeURL but allows more schemes (e.g. git, ssh, ftp)
// while still enforcing IP validation (blocking internal IPs/loopback) and blocking
// known dangerous schemes (gopher, expect, etc).
//
// This is suitable for validating URLs passed to command-line tools like curl/wget/git.
var IsSafeCommandURL = func(urlStr string) error {
	return validateURL(urlStr, true)
}

func validateURL(urlStr string, allowCommandSchemes bool) error {
	// Bypass if explicitly allowed (for testing/development)
	// We handle various truthy formats to be robust against YAML/Helm quoting issues
	envVal := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	cleanVal := strings.Trim(strings.ToLower(envVal), "\"' ")
	if cleanVal == "true" || cleanVal == "1" || cleanVal == "yes" || cleanVal == "on" {
		return nil
	}

	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == trueVal
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == trueVal

	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// 1. Check Scheme
	if allowCommandSchemes {
		// Block specifically dangerous schemes
		switch u.Scheme {
		case "file", "gopher", "dict", "expect", "php", "vbscript", "javascript", "data":
			return fmt.Errorf("unsafe scheme: %s", u.Scheme)
		}
		// Allow others, but enforce IP check below
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s (only http and https are allowed)", u.Scheme)
	}

	// 2. Resolve Host
	host := u.Hostname()
	if host == "" {
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
			// Debug log for CI troubleshooting (temporary)
			// fmt.Printf("SSRF Blocked: URL=%q, Host=%q, IP=%s, EnvVar=%q, CleanVar=%q\n", urlStr, host, ip.String(), envVal, cleanVal)
			return fmt.Errorf("ssrf attempt blocked: host %s resolved to loopback/private ip %s", host, ip.String())
		}
	}

	return nil
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
