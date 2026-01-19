// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

// SafeDialer is a dialer that validates the resolved IP address against
// security policies (e.g., prohibiting private IPs) before connecting.
type SafeDialer struct {
	dialer *net.Dialer
}

// DialContext resolves the address, validates the IPs, and connects to a safe IP.
// It effectively prevents DNS rebinding attacks by validating the resolved IP
// immediately before connection.
func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	// If network is not tcp/udp (e.g. unix), delegate to base dialer if we want to support it?
	// But usually we only want to validate TCP/UDP addresses.
	// Unix sockets are local files, so validation.IsAllowedPath should handle them if applicable,
	// but here we deal with network addresses.
	if network == "unix" || network == "unixpacket" || network == "unixgram" {
		return d.dialer.DialContext(ctx, network, addr)
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", addr, err)
	}

	// Resolve IPs
	resolver := d.dialer.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	ips, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %q: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for host %q", host)
	}

	// Validate IPs and find the first safe one
	var safeIP net.IP
	for _, ip := range ips {
		if err := ValidateIP(ip); err == nil {
			safeIP = ip
			break
		}
	}

	if safeIP == nil {
		// Log the first unsafe IP for debugging context in the error
		return nil, fmt.Errorf("host %q resolves to unsafe IPs (e.g., %s)", host, ips[0].String())
	}

	// Connect to the safe IP
	// We join the IP and port to form the address to dial.
	// IPv6 literals must be enclosed in brackets.
	targetAddr := net.JoinHostPort(safeIP.String(), port)

	return d.dialer.DialContext(ctx, network, targetAddr)
}

// NewSafeTransport returns an http.Transport configured to use SafeDialer.
// It uses recommended security defaults for timeouts.
func NewSafeTransport(tlsConfig *tls.Config) *http.Transport {
	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	safeDialer := &SafeDialer{
		dialer: baseDialer,
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSClientConfig:       tlsConfig,
		DialContext:           safeDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
