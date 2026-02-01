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
// Summary: Attaches a remote IP address to the context for downstream use.
//
// Parameters:
//   - ctx: context.Context. The parent context.
//   - ip: string. The remote IP address string.
//
// Returns:
//   - context.Context: A new context with the remote IP value associated with a private key.
//
// Throws/Errors:
//   - Does not panic.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// ExtractIP extracts and validates the IP address from a string.
//
// Summary: Parses and cleans an IP address string, handling ports, brackets, and zone indices.
//
// Parameters:
//   - addr: string. The raw address string (e.g., "192.168.1.1:80", "[::1]", "fe80::1%eth0").
//
// Returns:
//   - string: The canonical IP address string. Returns empty string if invalid.
//
// Throws/Errors:
//   - Returns empty string on parsing failure.
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
// Summary: Determines the client's IP address, optionally trusting proxy headers.
//
// Parameters:
//   - r: *http.Request. The HTTP request.
//   - trustProxy: bool. Whether to trust X-Real-IP and X-Forwarded-For headers.
//
// Returns:
//   - string: The detected client IP address, or empty string if not found.
//
// Throws/Errors:
//   - Does not panic.
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
// Summary: Retrieves the remote IP previously stored in the context.
//
// Parameters:
//   - ctx: context.Context. The context object.
//
// Returns:
//   - string: The remote IP address if found.
//   - bool: True if the IP was present in the context.
//
// Throws/Errors:
//   - Does not panic.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

// IsPrivateNetworkIP checks if the IP address belongs to a private network.
//
// Summary: Checks if an IP is a private network address (RFC1918, RFC4193, RFC6598).
//
// Parameters:
//   - ip: net.IP. The IP address to check.
//
// Returns:
//   - bool: True if the IP is private (excluding loopback/link-local).
//
// Throws/Errors:
//   - Does not panic.
func IsPrivateNetworkIP(ip net.IP) bool {
	return validation.IsPrivateNetworkIP(ip)
}

// IsPrivateIP checks if the IP address is private, link-local, or loopback.
//
// Summary: Checks if an IP is private, link-local, or loopback.
//
// Parameters:
//   - ip: net.IP. The IP address to check.
//
// Returns:
//   - bool: True if the IP is not publicly routable.
//
// Throws/Errors:
//   - Does not panic.
func IsPrivateIP(ip net.IP) bool {
	return validation.IsPrivateIP(ip)
}

func isNAT64Loopback(ip net.IP) bool {
	return validation.IsNAT64Loopback(ip)
}

func isNAT64LinkLocal(ip net.IP) bool {
	return validation.IsNAT64LinkLocal(ip)
}
