// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"fmt"
	"net"
	"os"
)

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
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port: %w", err)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("dns lookup failed for host %s: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no ip addresses found for host: %s", host)
	}

	// Check all resolved IPs. If any are forbidden, block the request.
	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK") == "true"
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORKS") == "true"

	for _, ip := range ips {
		if !allowLoopback && (ip.IsLoopback() || ip.IsLinkLocalUnicast()) {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to loopback/link-local ip %s", host, ip)
		}
		if !allowPrivate && ip.IsPrivate() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to private ip %s", host, ip)
		}
		if ip.IsUnspecified() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to unspecified ip %s", host, ip)
		}
	}

	// All IPs are safe. Try connecting to them in order.
	var lastErr error
	for _, ip := range ips {
		dialAddr := net.JoinHostPort(ip.String(), port)
		conn, err := (&net.Dialer{}).DialContext(ctx, network, dialAddr)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("failed to connect to any resolved IP: %w", lastErr)
}
