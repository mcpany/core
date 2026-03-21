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

// Summary: ContextWithRemoteIP creates a new context containing the remote IP address. Injects the remote IP into the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - ip (string): The ip parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// Summary: ExtractIP extracts and validates the IP address from a string. Parses and sanitizes an IP address string. It handles "host:port" formats, strips IPv6 brackets, and removes zone indices.
//
// Parameters:
//   - addr (string): The addr parameter.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: GetClientIP extracts the client IP address from an HTTP request. Determines the client's IP address.
//
// Parameters:
//   - r (*http.Request): The r parameter.
//   - trustProxy (bool): The trustProxy parameter.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: RemoteIPFromContext retrieves the remote IP address stored in the context. Retrieves the remote IP from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

// Summary: IsPrivateNetworkIP checks if the IP address belongs to a private network. Checks if an IP is a private network address. This includes RFC1918 (Private IPv4), RFC4193 (Unique Local IPv6), and RFC6598 (CGNAT). It does NOT include loopback or link-local addresses.
//
// Parameters:
//   - ip (net.IP): The ip parameter.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func IsPrivateNetworkIP(ip net.IP) bool {
	return validation.IsPrivateNetworkIP(ip)
}

// Summary: IsPrivateIP checks if the IP address is private, link-local, or loopback. Checks if an IP is internal/private. This is a comprehensive check for any "internal" IP address that shouldn't be publicly routable.
//
// Parameters:
//   - ip (net.IP): The ip parameter.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func IsPrivateIP(ip net.IP) bool {
	return validation.IsPrivateIP(ip)
}

func isNAT64Loopback(ip net.IP) bool {
	return validation.IsNAT64Loopback(ip)
}

func isNAT64LinkLocal(ip net.IP) bool {
	return validation.IsNAT64LinkLocal(ip)
}
