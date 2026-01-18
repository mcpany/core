// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package util provides network and other utility functions.
package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// IPResolver defines an interface for looking up IP addresses.
// It matches net.Resolver.LookupIP signature.
type IPResolver interface {
	// LookupIP looks up host using the local resolver.
	// It returns a slice of that host's IPv4 and IPv6 addresses.
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// NetDialer defines an interface for dialing network connections.
// It matches net.Dialer.DialContext signature.
type NetDialer interface {
	// DialContext connects to the address on the named network using
	// the provided context.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// SafeDialer provides control over outbound connections to prevent SSRF.
type SafeDialer struct {
	// AllowLoopback allows connections to loopback addresses (127.0.0.1, ::1).
	AllowLoopback bool
	// AllowPrivate allows connections to private network addresses (RFC 1918, RFC 4193).
	AllowPrivate bool
	// AllowLinkLocal allows connections to link-local addresses (169.254.x.x, fe80::/10).
	// This includes cloud metadata services.
	AllowLinkLocal bool

	// Resolver is used for DNS lookups. If nil, net.DefaultResolver is used.
	Resolver IPResolver
	// Dialer is used for establishing connections. If nil, &net.Dialer{} is used.
	Dialer NetDialer
}

// NewSafeDialer creates a new SafeDialer with strict defaults (blocking all non-public IPs).
//
// Returns the result.
func NewSafeDialer() *SafeDialer {
	return &SafeDialer{
		AllowLoopback:  false,
		AllowPrivate:   false,
		AllowLinkLocal: false,
	}
}

// DialContext creates a connection to the given address, enforcing the configured egress policy.
//
// ctx is the context for the request.
// network is the network.
// addr is the addr.
//
// Returns the result.
// Returns an error if the operation fails.
func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port: %w", err)
	}

	resolver := d.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	lookupNetwork := "ip"
	if strings.HasSuffix(network, "4") {
		lookupNetwork = "ip4"
	} else if strings.HasSuffix(network, "6") {
		lookupNetwork = "ip6"
	}

	ips, err := resolver.LookupIP(ctx, lookupNetwork, host)
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
		if !d.AllowPrivate && IsPrivateNetworkIP(ip) {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to private ip %s", host, ip)
		}
	}

	dialer := d.Dialer
	if dialer == nil {
		dialer = &net.Dialer{}
	}

	// All IPs are safe. Dial them in order until one succeeds.
	var firstErr error
	for _, ip := range ips {
		dialAddr := net.JoinHostPort(ip.String(), port)
		conn, err := dialer.DialContext(ctx, network, dialAddr)
		if err == nil {
			return conn, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return nil, firstErr
}

// SafeDialContext creates a connection to the given address, preventing
// connections to private or loopback IP addresses to mitigate SSRF vulnerabilities.
// It respects the MCPANY_ALLOW_LOOPBACK_RESOURCES and MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES
// environment variables.
//
// ctx is the context for the dial operation.
// network is the network type (e.g., "tcp").
// addr is the address to connect to (host:port).
// It returns the established connection or an error if the connection fails or is blocked.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := NewSafeDialer()
	if os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == TrueStr {
		dialer.AllowLoopback = true
	}
	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == TrueStr {
		dialer.AllowPrivate = true
	}
	return dialer.DialContext(ctx, network, addr)
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

// CheckConnection attempts to establish a TCP connection to the given address.
// This is useful for verifying if a service is reachable.
func CheckConnection(ctx context.Context, address string) error {
	var target string
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return fmt.Errorf("failed to parse address %s: %w", address, err)
		}
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
		target = net.JoinHostPort(host, port)
	} else {
		// If no scheme, try to parse as host:port. If no port, assume 80.
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			// If SplitHostPort fails, it means no port was specified.
			// Assume it's just a hostname and default to port 80.
			host = address
			port = "80"
		}
		target = net.JoinHostPort(host, port)
	}

	conn, err := SafeDialContext(ctx, "tcp", target)
	if err != nil {
		return fmt.Errorf("failed to connect to address %s: %w", target, err)
	}
	defer func() { _ = conn.Close() }()
	return nil
}
