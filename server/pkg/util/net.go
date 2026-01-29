// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package util provides network and other utility functions.
package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// IPResolver defines an interface for looking up IP addresses.
// It matches the signature of net.Resolver.LookupIP.
type IPResolver interface {
	// LookupIP looks up host using the local resolver.
	// It returns a slice of that host's IPv4 and IPv6 addresses.
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// NetDialer defines an interface for dialing network connections.
// It matches the signature of net.Dialer.DialContext.
type NetDialer interface {
	// DialContext connects to the address on the named network using
	// the provided context.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// SafeDialer provides control over outbound connections to prevent Server-Side Request Forgery (SSRF).
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

// NewSafeDialer creates a new SafeDialer with strict default security settings.
// By default, it blocks all non-public IP addresses (loopback, private, link-local).
//
// Returns:
//   - (*SafeDialer): A new SafeDialer instance with restrictive defaults.
func NewSafeDialer() *SafeDialer {
	return &SafeDialer{
		AllowLoopback:  false,
		AllowPrivate:   false,
		AllowLinkLocal: false,
	}
}

// DialContext establishes a network connection to the given address while enforcing egress policies.
// It resolves the host's IP addresses and verifies them against the allowed list before connecting.
//
// Parameters:
//   - ctx (context.Context): The context for the dial operation.
//   - network (string): The network type (e.g., "tcp", "tcp4", "tcp6").
//   - addr (string): The address to connect to (host:port).
//
// Returns:
//   - (net.Conn): The established connection.
//   - (error): An error if resolution fails, all resolved IPs are blocked by policy, or the connection fails.
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
		if !d.AllowLoopback && (ip.IsLoopback() || isNAT64Loopback(ip) || ip.IsUnspecified()) {
			return nil, fmt.Errorf("ssrf attempt blocked: host %s resolved to loopback ip %s", host, ip)
		}
		if !d.AllowLinkLocal && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || isNAT64LinkLocal(ip)) {
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

// SafeDialContext creates a connection to the given address with strict SSRF protection.
// It is a convenience wrapper around SafeDialer with default settings (blocking private/loopback).
//
// Parameters:
//   - ctx (context.Context): The context for the dial operation.
//   - network (string): The network type.
//   - addr (string): The address to connect to (host:port).
//
// Returns:
//   - (net.Conn): The established connection.
//   - (error): An error if the connection is blocked by policy or fails.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return NewSafeDialer().DialContext(ctx, network, addr)
}

// NewSafeHTTPClient creates a new HTTP client configured to prevent SSRF attacks.
// It uses a custom Transport backed by SafeDialer.
//
// Configuration is loaded from environment variables:
//   - MCPANY_ALLOW_LOOPBACK_RESOURCES: Set to "true" to allow loopback connections.
//   - MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES: Set to "true" to allow private network connections.
//
// Returns:
//   - (*http.Client): A configured HTTP client.
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

// CheckConnection verifies if a TCP connection can be established to the given address.
// This is typically used for health checks or validating upstream service reachability.
// It uses SafeDialer to respect egress policies, but allows overriding via environment variables.
//
// Parameters:
//   - ctx (context.Context): The context for the connection attempt.
//   - address (string): The target address (URL or host:port).
//
// Returns:
//   - (error): nil if the connection succeeded, or an error if it failed.
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
			// If host is an IPv6 literal with brackets, strip them because JoinHostPort adds them back.
			if len(host) > 0 && host[0] == '[' && host[len(host)-1] == ']' {
				host = host[1 : len(host)-1]
			}
			port = "80"
		}
		target = net.JoinHostPort(host, port)
	}

	// Use SafeDialer to prevent SSRF during connectivity checks
	dialer := NewSafeDialer()
	// Allow overriding safety checks via environment variables (consistent with validation package)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == TrueStr {
		dialer.AllowLoopback = true
		dialer.AllowPrivate = true
	}

	if os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == TrueStr {
		dialer.AllowLoopback = true
	}

	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == TrueStr {
		dialer.AllowPrivate = true
	}

	dialer.Dialer = &net.Dialer{Timeout: 5 * time.Second}

	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return fmt.Errorf("failed to connect to address %s: %w", target, err)
	}
	defer func() { _ = conn.Close() }()
	return nil
}

// ListenWithRetry attempts to listen on the given address with retries to handle transient port conflicts.
// It is particularly useful for avoiding race conditions when binding to port 0 (dynamic allocation)
// in high-churn environments.
//
// Parameters:
//   - ctx (context.Context): The context for the listen operation.
//   - network (string): The network type (e.g., "tcp").
//   - address (string): The address to listen on.
//
// Returns:
//   - (net.Listener): The successfully bound listener.
//   - (error): An error if binding fails after all retries.
func ListenWithRetry(ctx context.Context, network, address string) (net.Listener, error) {
	var lis net.Listener
	var err error

	// We try 10 times for port 0 to mitigate OS-level races.
	maxRetries := 1
	if strings.HasSuffix(address, ":0") {
		maxRetries = 10
	}

	for i := 0; i < maxRetries; i++ {
		lis, err = (&net.ListenConfig{}).Listen(ctx, network, address)
		if err == nil {
			return lis, nil
		}

		// Check if the error is "address already in use" or similar EADDRINUSE indicator.
		errStr := strings.ToLower(err.Error())
		isBindErr := strings.Contains(errStr, "address already in use") ||
			strings.Contains(errStr, "eaddrinuse")

		// If it's not a bind error, or we're at the last attempt, return.
		if !isBindErr || i == maxRetries-1 {
			return nil, err
		}

		// Exponential backoff with jitter: 100ms, 200ms, 400ms, 800ms, 1.6s...
		// We start slightly higher than before (100ms) to give more room.
		backoff := time.Duration(100*math.Pow(2, float64(i))) * time.Millisecond
		// Add jitter (up to 50ms)
		jitterBig, err := rand.Int(rand.Reader, big.NewInt(50))
		var jitter int64
		if err == nil {
			jitter = jitterBig.Int64()
		}
		backoff += time.Duration(jitter) * time.Millisecond

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			// retry
		}
	}
	return nil, err
}
