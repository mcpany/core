// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package util provides network and other utility functions.
package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// SafeDialer provides control over outbound connections to prevent SSRF.
type SafeDialer struct {
	// AllowLoopback allows connections to loopback addresses (127.0.0.1, ::1).
	AllowLoopback bool
	// AllowPrivate allows connections to private network addresses (RFC 1918, RFC 4193).
	AllowPrivate bool
	// AllowLinkLocal allows connections to link-local addresses (169.254.x.x, fe80::/10).
	// This includes cloud metadata services.
	AllowLinkLocal bool
}

// NewSafeDialer creates a new SafeDialer with strict defaults (blocking all non-public IPs).
func NewSafeDialer() *SafeDialer {
	return &SafeDialer{
		AllowLoopback:  false,
		AllowPrivate:   false,
		AllowLinkLocal: false,
	}
}

// DialContext creates a connection to the given address, enforcing the configured egress policy.
func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port: %w", err)
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("dns lookup failed for host %s: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no ip addresses found for host: %s", host)
	}

	// Check all resolved IPs. If any are forbidden, block the request.
	for _, ip := range ips {
		if !d.AllowLoopback && ip.IsLoopback() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to loopback ip %s", host, ip)
		}
		if !d.AllowLinkLocal && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to link-local ip %s", host, ip)
		}
		if !d.AllowPrivate && ip.IsPrivate() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to private ip %s", host, ip)
		}
		if ip.IsUnspecified() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to unspecified ip %s", host, ip)
		}
	}

	// All IPs are safe. Dial the first one.
	dialAddr := net.JoinHostPort(ips[0].String(), port)
	return (&net.Dialer{}).DialContext(ctx, network, dialAddr)
}

// SafeDialContext creates a connection to the given address, but strictly prevents
// connections to private or loopback IP addresses to mitigate SSRF vulnerabilities.
//
// It resolves the host's IP addresses and checks each one. If any resolved IP
// is private or loopback, the connection is blocked.
//
// ctx is the context for the dial operation.
// network is the network type (e.g., "tcp").
// addr is the address to connect to (host:port).
// It returns the established connection or an error if the connection fails or is blocked.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return NewSafeDialer().DialContext(ctx, network, addr)
}

// NewSafeHTTPClient creates a new http.Client that prevents SSRF.
// It uses SafeDialer to block access to link-local and private IPs.
// Configuration can be overridden via environment variables:
// - MCPANY_ALLOW_LOOPBACK_RESOURCES: Set to "true" to allow loopback addresses.
// - MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES: Set to "true" to allow private network addresses.
func NewSafeHTTPClient() *http.Client {
	dialer := NewSafeDialer()
	if os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == TrueStr {
		dialer.AllowLoopback = true
	}
	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == TrueStr {
		dialer.AllowPrivate = true
	}
	// LinkLocal is always blocked by default and cannot be enabled via env var for now (safest default).

	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
}
