// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/validation"
)

type contextKey string

const remoteIPContextKey = contextKey("remote_ip")

// ContextWithRemoteIP creates a new context containing the remote IP address.
//
// Summary: Creates a new context containing the remote IP address.
//
// Parameters:
//   - ctx: context.Context. The parent context.
//   - ip: string. The remote IP address to store in the context.
//
// Returns:
//   - context.Context: A new context with the remote IP attached.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// ExtractIP extracts and validates the IP address from a string.
//
// Summary: Extracts and validates the IP address from a string.
//
// Parameters:
//   - addr: string. The address string to parse (e.g., "192.168.1.1:80", "[::1]", "fe80::1%eth0").
//
// Returns:
//   - string: The cleaned IP address string, or an empty string if the address is invalid.
func ExtractIP(addr string) string {
	ipStr, _, err := net.SplitHostPort(addr)
	if err != nil {
		ipStr = addr
	}
	if len(ipStr) > 0 && ipStr[0] == '[' && ipStr[len(ipStr)-1] == ']' {
		ipStr = ipStr[1 : len(ipStr)-1]
	}
	// Strip zone index if present (e.g. fe80::1%eth0 -> fe80::1)
	if idx := strings.IndexByte(ipStr, '%'); idx != -1 {
		ipStr = ipStr[:idx]
	}

	// Validate IP
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return ""
	}
	return parsedIP.String()
}

// GetClientIP extracts the client IP address from an HTTP request.
//
// Summary: Extracts the client IP address from an HTTP request.
//
// Parameters:
//   - r: *http.Request. The HTTP request to inspect.
//   - trustProxy: bool. If true, trusts 'X-Real-IP' and 'X-Forwarded-For' headers. If false, only uses 'RemoteAddr'.
//
// Returns:
//   - string: The best-effort client IP address.
func GetClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		// Prefer X-Real-IP as it is usually a single IP set by the trusted proxy.
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			if ip := ExtractIP(xri); ip != "" {
				return ip
			}
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// Use the first IP in the list (client IP)
			// Optimization: Use strings.Cut to avoid allocating a slice for all parts
			// in case of multiple IPs in the header.
			clientIP, _, _ := strings.Cut(xff, ",")
			clientIP = strings.TrimSpace(clientIP)
			if clientIP != "" {
				if ip := ExtractIP(clientIP); ip != "" {
					return ip
				}
			}
		}
	}

	// Fallback to RemoteAddr
	return ExtractIP(r.RemoteAddr)
}

// RemoteIPFromContext retrieves the remote IP address stored in the context.
//
// Summary: Retrieves the remote IP address stored in the context.
//
// Parameters:
//   - ctx: context.Context. The context to retrieve the IP from.
//
// Returns:
//   - string: The remote IP address.
//   - bool: True if the IP was found, false otherwise.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

// IsPrivateNetworkIP checks if the IP address belongs to a private network.
//
// Summary: Checks if the IP address belongs to a private network.
//
// This includes RFC1918 (Private IPv4), RFC4193 (Unique Local IPv6), and RFC6598 (CGNAT).
// It does NOT include loopback or link-local addresses.
//
// Parameters:
//   - ip: net.IP. The IP address to check.
//
// Returns:
//   - bool: True if the IP is a private network address.
func IsPrivateNetworkIP(ip net.IP) bool {
	return validation.IsPrivateNetworkIP(ip)
}

// IsPrivateIP checks if the IP address is private, link-local, or loopback.
//
// Summary: Checks if the IP address is private, link-local, or loopback.
//
// This is a comprehensive check for any "internal" IP address that shouldn't be publicly routable.
//
// Parameters:
//   - ip: net.IP. The IP address to check.
//
// Returns:
//   - bool: True if the IP is private, link-local, or loopback.
func IsPrivateIP(ip net.IP) bool {
	return validation.IsPrivateIP(ip)
}

func isNAT64Loopback(ip net.IP) bool {
	return validation.IsNAT64Loopback(ip)
}

func isNAT64LinkLocal(ip net.IP) bool {
	return validation.IsNAT64LinkLocal(ip)
}
