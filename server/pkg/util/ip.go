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
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}
	if len(ip) > 0 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		ip = ip[1 : len(ip)-1]
	}
	// Strip zone index if present (e.g. fe80::1%eth0 -> fe80::1)
	for i := 0; i < len(ip); i++ {
		if ip[i] == '%' {
			return ip[:i]
		}
	}
	return ip
}

// GetClientIP extracts the client IP from the request.
// If trustProxy is true, it respects X-Real-IP and X-Forwarded-For headers.
func GetClientIP(r *http.Request, trustProxy bool) string {
	ip := ExtractIP(r.RemoteAddr)

	if trustProxy {
		// Prefer X-Real-IP as it is usually a single IP set by the trusted proxy.
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return ExtractIP(xri)
		}
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

// IsLoopback checks if the IP address is a loopback address.
// This includes IPv4 127.0.0.0/8, IPv6 ::1, IPv4-compatible loopback, and NAT64 loopback.
//
// ip is the ip.
//
// Returns true if successful.
func IsLoopback(ip net.IP) bool {
	return validation.IsLoopback(ip)
}
