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

// ContextWithRemoteIP provides contextwithremoteip functionality.
//
// Summary: ContextWithRemoteIP.
//
// Parameters.
//   - ctx: The parameter.
//   - ip: The parameter.
//
// Returns.
//   - result: The result.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// ExtractIP provides extractip functionality.
//
// Summary: ExtractIP.
//
// Parameters.
//   - addr: The parameter.
//
// Returns.
//   - result: The result.
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

// GetClientIP provides getclientip functionality.
//
// Summary: GetClientIP.
//
// Parameters.
//   - r: The parameter.
//   - trustProxy: The parameter.
//
// Returns.
//   - result: The result.
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

// RemoteIPFromContext provides remoteipfromcontext functionality.
//
// Summary: RemoteIPFromContext.
//
// Parameters.
//   - ctx: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

// IsPrivateNetworkIP provides isprivatenetworkip functionality.
//
// Summary: IsPrivateNetworkIP.
//
// Parameters.
//   - ip: The parameter.
//
// Returns.
//   - result: The result.
func IsPrivateNetworkIP(ip net.IP) bool {
	return validation.IsPrivateNetworkIP(ip)
}

// IsPrivateIP provides isprivateip functionality.
//
// Summary: IsPrivateIP.
//
// Parameters.
//   - ip: The parameter.
//
// Returns.
//   - result: The result.
func IsPrivateIP(ip net.IP) bool {
	return validation.IsPrivateIP(ip)
}

func isNAT64Loopback(ip net.IP) bool {
	return validation.IsNAT64Loopback(ip)
}

func isNAT64LinkLocal(ip net.IP) bool {
	return validation.IsNAT64LinkLocal(ip)
}
