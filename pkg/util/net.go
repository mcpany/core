// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"fmt"
	"net"
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
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() || ip.IsUnspecified() {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to a private ip %s", host, ip)
		}
	}

	// All IPs are safe. Dial the first one.
	dialAddr := net.JoinHostPort(ips[0].String(), port)
	return (&net.Dialer{}).DialContext(ctx, network, dialAddr)
}
