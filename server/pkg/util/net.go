// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

// SafeDialContext creates a connection to the given address, but strictly prevents
// connections to private or loopback IP addresses to mitigate SSRF vulnerabilities.
//
// It uses net.Dialer's Control hook to validate the IP address before connection,
// preserving DNS load balancing and Happy Eyeballs behavior.
//
// To allow connections to loopback or private networks, set the environment variables
// MCPANY_ALLOW_LOOPBACK=true or MCPANY_ALLOW_PRIVATE_NETWORK=true respectively.
//
// ctx is the context for the dial operation.
// network is the network type (e.g., "tcp").
// addr is the address to connect to (host:port).
// It returns the established connection or an error if the connection fails or is blocked.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK") == "true"
	allowPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK") == "true"

	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return fmt.Errorf("failed to split host port in control: %w", err)
			}
			ip := net.ParseIP(host)
			if ip == nil {
				return fmt.Errorf("failed to parse ip in control: %s", host)
			}

			// Always block link-local unicast and unspecified (0.0.0.0) as they are rarely valid for egress
			// and often indicate misconfiguration or attack.
			if ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
				return fmt.Errorf("ssrf attempt blocked: restricted ip %s", ip)
			}

			if !allowLoopback && ip.IsLoopback() {
				return fmt.Errorf("ssrf attempt blocked: loopback ip %s", ip)
			}

			if !allowPrivate && ip.IsPrivate() {
				return fmt.Errorf("ssrf attempt blocked: private ip %s", ip)
			}
			return nil
		},
	}

	return dialer.DialContext(ctx, network, addr)
}
