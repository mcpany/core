// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"context"
	"net"
)

type contextKey string

const remoteIPContextKey = contextKey("remote_ip")

// ContextWithRemoteIP returns a new context with the remote IP.
func ContextWithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, remoteIPContextKey, ip)
}

// ExtractIP extracts the IP address from a host:port string or just an IP string.
// It also handles IPv6 brackets.
func ExtractIP(addr string) string {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}
	if len(ip) > 0 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		ip = ip[1 : len(ip)-1]
	}
	return ip
}

// RemoteIPFromContext retrieves the remote IP from the context.
func RemoteIPFromContext(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(remoteIPContextKey).(string)
	return ip, ok
}

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"fc00::/7",       // RFC4193 unique local address
		"fe80::/10",      // RFC4291 link-local
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateIPBlocks = append(privateIPBlocks, block)
		}
	}
}

// IsPrivateIP checks if the IP address is a private or link-local address.
func IsPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
