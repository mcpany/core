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
func ExtractIP(addr string) string {
	if len(addr) == 0 {
		return ""
	}

	// ⚡ Bolt Optimization: Fast path to avoid net.SplitHostPort allocation (AddrError).
	// We only call SplitHostPort if we detect a port is likely present.
	needsSplit := true
	firstColon := strings.IndexByte(addr, ':')

	switch {
	case firstColon == -1:
		// No colon -> No port (e.g. "1.2.3.4", "localhost")
		needsSplit = false
	case addr[0] == '[':
		// Starts with bracket
		// If it ends with bracket, it's just "[IPv6]" (no port)
		if addr[len(addr)-1] == ']' {
			needsSplit = false
		}
	default:
		// No brackets, but has colon.
		// If multiple colons -> IPv6 literal (e.g. "::1") -> No port (SplitHostPort would fail)
		// If single colon -> IPv4 with port (e.g. "1.2.3.4:80")
		if strings.LastIndexByte(addr, ':') != firstColon {
			needsSplit = false
		}
	}

	var ip string
	if needsSplit {
		var err error
		var host string
		host, _, err = net.SplitHostPort(addr)
		if err == nil {
			ip = host
		} else {
			ip = addr
		}
	} else {
		ip = addr
	}

	// Remove brackets if present (e.g. if we skipped SplitHostPort for "[::1]")
	if len(ip) > 1 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		ip = ip[1 : len(ip)-1]
	}

	// Strip zone index if present (e.g. fe80::1%eth0 -> fe80::1)
	if idx := strings.IndexByte(ip, '%'); idx != -1 {
		return ip[:idx]
	}
	return ip
}

// GetClientIP extracts the client IP from the request.
// If trustProxy is true, it respects X-Forwarded-For header.
func GetClientIP(r *http.Request, trustProxy bool) string {
	ip := ExtractIP(r.RemoteAddr)

	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// Use the first IP in the list (client IP)
			// Optimization: Use strings.Cut to avoid allocating a slice for all parts
			// in case of multiple IPs in the header.
			clientIP, _, _ := strings.Cut(xff, ",")
			clientIP = strings.TrimSpace(clientIP)
			if clientIP != "" {
				ip = ExtractIP(clientIP)
			}
		}
	}
	return ip
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
