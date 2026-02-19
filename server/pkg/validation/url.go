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

// LookupIPFunc is a variable to allow mocking DNS resolution in tests.
var LookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, network, host)
}

// IsSafeHost checks if the host is safe to connect to.
// It resolves the host to ensure it doesn't point to
// loopback, link-local, private, or multicast addresses.
var IsSafeHost = func(host string) error {
	// Bypass if explicitly allowed (for testing/development)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == trueVal {
		return nil
	}

	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == trueVal
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == trueVal

	if host == "" {
		return fmt.Errorf("missing host")
	}

	// Check if host is an IP literal
	if ip := net.ParseIP(host); ip != nil {
		return validateIP(ip, allowLoopback, allowPrivate)
	}

	// Resolve Domain
	// Use a timeout for resolution
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := LookupIPFunc(ctx, "ip", host)
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

	// 2. Resolve Host
	return IsSafeHost(u.Hostname())
}

// IsSafeGitURL checks if the URL is safe for Git operations.
// It supports ssh://, git://, http://, https:// and SCP-style syntax.
// It validates the host to ensure it doesn't point to unsafe addresses.
var IsSafeGitURL = func(urlStr string) error {
	// 1. Check for standard URL syntax
	if strings.Contains(urlStr, "://") {
		u, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}

		// Supported schemes for Git
		switch u.Scheme {
		case "http", "https", "ssh", "git":
			// Allowed
		default:
			return fmt.Errorf("unsupported scheme for git: %s", u.Scheme)
		}

		return IsSafeHost(u.Hostname())
	}

	// 2. Check for SCP-style syntax: [user@]host:path
	idx := strings.Index(urlStr, ":")
	if idx == -1 {
		// Not SCP syntax (no colon). Treated as local path.
		return nil
	}

	// Git rule: if slash exists before colon, it's a local path, not SCP.
	slashIdx := strings.Index(urlStr, "/")
	if slashIdx != -1 && slashIdx < idx {
		return nil
	}

	// Also check backslash for Windows paths compatibility?
	// Git on Linux doesn't treat backslash as separator, but for robustness:
	backSlashIdx := strings.Index(urlStr, "\\")
	if backSlashIdx != -1 && backSlashIdx < idx {
		return nil
	}

	// Extract host part
	hostPart := urlStr[:idx]
	// If user@ present, strip it
	if atIdx := strings.Index(hostPart, "@"); atIdx != -1 {
		hostPart = hostPart[atIdx+1:]
	}

	// hostPart cannot be empty
	if hostPart == "" {
		// :path is not valid SCP syntax usually, maybe treated as local file?
		return nil
	}

	// Validate hostPart
	// If the host part looks like a drive letter (single char), we might be careful.
	// But as reasoned, IsSafeHost will handle it.

	return IsSafeHost(hostPart)
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
