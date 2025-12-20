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
// It uses net.Dialer's Control hook to validate IP addresses after resolution but
// before connection, ensuring Happy Eyeballs support.
//
// Security Rules:
// - Loopback IPs (127.0.0.0/8, ::1): Blocked by default. Allow with MCPANY_ALLOW_LOOPBACK=true.
// - Link-Local IPs (169.254.0.0/16, fe80::/10): Always Blocked (protects cloud metadata services).
// - Unspecified IPs (0.0.0.0, ::): Always Blocked.
// - Private IPs (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16): ALLOWED by default. Block with MCPANY_ALLOW_PRIVATE_NETWORK=false.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	allowLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK") == "true"

	// Default to allowing private networks to avoid breaking changes for upstream services
	// running in internal networks (e.g. K8s, VPCs).
	// Users can strict block private networks by setting MCPANY_ALLOW_PRIVATE_NETWORK=false.
	allowPrivateNetwork := true
	if val := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK"); val != "" {
		allowPrivateNetwork = val == "true"
	}

	d := &net.Dialer{
		Timeout: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			// address is IP:Port
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return fmt.Errorf("failed to split host/port in control: %w", err)
			}
			ip := net.ParseIP(host)
			if ip == nil {
				return fmt.Errorf("failed to parse IP in control: %s", host)
			}

			if ip.IsUnspecified() {
				return fmt.Errorf("ssrf attempt blocked: unspecified ip %s", ip)
			}
			if ip.IsLinkLocalUnicast() {
				return fmt.Errorf("ssrf attempt blocked: link-local ip %s", ip)
			}
			if !allowLoopback && ip.IsLoopback() {
				return fmt.Errorf("ssrf attempt blocked: loopback ip %s", ip)
			}
			if !allowPrivateNetwork && ip.IsPrivate() {
				return fmt.Errorf("ssrf attempt blocked: private ip %s", ip)
			}

			return nil
		},
	}
	return d.DialContext(ctx, network, addr)
}
