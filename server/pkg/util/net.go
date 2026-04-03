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

// IPResolver represents the public IPResolver entity.
//
// Summary: Defines the structured data model representing a resolver.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type IPResolver interface {
	// LookupIP looks up host using the local resolver.
	//
	// Summary: Looks up host IP addresses.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the lookup.
	//   - network (string): The network type (e.g., "ip", "ip4", "ip6").
	//   - host (string): The host to look up.
	//
	// Returns:
	//   - ([]net.IP): A slice of that host's IPv4 and IPv6 addresses.
	//   - (error): An error if the lookup fails.
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// NetDialer represents the public NetDialer entity.
//
// Summary: Defines the structured data model representing a dialer.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type NetDialer interface {
	// DialContext connects to the address on the named network using the provided context.
	//
	// Summary: Dials a network connection.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the dial operation.
	//   - network (string): The network type.
	//   - address (string): The address to connect to.
	//
	// Returns:
	//   - (net.Conn): The established connection.
	//   - (error): An error if the connection fails.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// SafeDialer represents the public SafeDialer entity.
//
// Summary: Defines the structured data model representing a dialer.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// NewSafeDialer serves as a public interface for interacting with NewSafeDialer.
//
// Summary: Constructs and returns an initialized safe dialer ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewSafeDialer() *SafeDialer {
	return &SafeDialer{
		AllowLoopback:  false,
		AllowPrivate:   false,
		AllowLinkLocal: false,
	}
}

// DialContext serves as a public interface for interacting with DialContext.
//
// Summary: Dial the context appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// SafeDialContext serves as a public interface for interacting with SafeDialContext.
//
// Summary: Safe the dial context appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return NewSafeDialer().DialContext(ctx, network, addr)
}

// NewSafeHTTPClient serves as a public interface for interacting with NewSafeHTTPClient.
//
// Summary: Constructs and returns an initialized safe http client ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewSafeHTTPClient() *http.Client {
	dialer := NewSafeDialer()
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
	// LinkLocal is always blocked by default and cannot be enabled via env var for now (safest default).

	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
}

// CheckConnection serves as a public interface for interacting with CheckConnection.
//
// Summary: Check the connection appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ListenWithRetry serves as a public interface for interacting with ListenWithRetry.
//
// Summary: Listen the with retry appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
