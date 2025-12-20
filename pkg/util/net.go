// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"fmt"
	"net"
)

// SafeDialer is a wrapper around net.Dialer that prevents SSRF attacks.
type SafeDialer struct {
	net.Dialer
	AllowLoopback bool
	AllowPrivate  bool
}

// DialContext connects to the address on the named network.
func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port: %w", err)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("dns lookup failed for host %s: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for host: %s", host)
	}

	// Check all resolved IPs. If any are forbidden, block the request.
	for _, ip := range ips {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to a link-local IP %s", host, ip)
		}
		if !d.AllowLoopback && (ip.IsLoopback() || ip.IsUnspecified()) {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to a loopback/unspecified IP %s", host, ip)
		}
		if !d.AllowPrivate && ip.IsPrivate() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to a private IP %s", host, ip)
		}
	}

	// All IPs are safe. Dial the first one.
	dialAddr := net.JoinHostPort(ips[0].String(), port)
	return d.Dialer.DialContext(ctx, network, dialAddr)
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
	d := &SafeDialer{
		AllowLoopback: false,
		AllowPrivate:  false,
	}
	return d.DialContext(ctx, network, addr)
}
