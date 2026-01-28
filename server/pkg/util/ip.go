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

// ContextWithRemoteIP returns a new context with the remote IP.
//
// ctx is the context for the request.
// ip is the ip.
//
// Returns the result.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// ExtractIP extracts the IP address from a host:port string or just an IP string.
// It also handles IPv6 brackets and strips IPv6 zone indices (e.g., %eth0).
// It returns an empty string if the extracted IP is invalid.
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

// GetClientIP extracts the client IP from the request.
// If trustProxy is true, it respects X-Real-IP and X-Forwarded-For headers.
// It validates that the extracted IP is a valid IP address.
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

// RemoteIPFromContext retrieves the remote IP from the context.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns true if successful.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

// IsPrivateNetworkIP checks if the IP address is a private network address.
// This includes RFC1918, RFC4193 (Unique Local), and RFC6598 (CGNAT).
// It does NOT include loopback or link-local addresses.
func IsPrivateNetworkIP(ip net.IP) bool {
	return validation.IsPrivateNetworkIP(ip)
}

// IsPrivateIP checks if the IP address is a private, link-local, or loopback address.
//
// ip is the ip.
//
// Returns true if successful.
func IsPrivateIP(ip net.IP) bool {
	return validation.IsPrivateIP(ip)
}

func isNAT64Loopback(ip net.IP) bool {
	return validation.IsNAT64Loopback(ip)
}

func isNAT64LinkLocal(ip net.IP) bool {
	return validation.IsNAT64LinkLocal(ip)
}

// IsLoopback checks if the IP address is a loopback address.
func IsLoopback(ip net.IP) bool {
	return validation.IsLoopback(ip)
}

// IsLinkLocal checks if the IP address is a link-local address.
func IsLinkLocal(ip net.IP) bool {
	return validation.IsLinkLocal(ip)
}
