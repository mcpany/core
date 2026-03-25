// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//revive:disable:var-naming
package util

import (
	"context"
// ContextWithRemoteIP creates a new context containing the remote IP address.
//
// ExtractIP extracts and validates the IP address from a string.
//
// Summary: Parses and sanitizes an IP address string.
//
// It handles "host:port" formats, strips IPv6 brackets, and removes zone indices.
//
// Parameters:
//   - addr: string. The address string to parse (e.g., "192.168.1.1:80", "[::1]", "fe80::1%eth0").
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - string: The cleaned IP address string, or an empty string if the address is invalid.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// GetClientIP extracts the client IP address from an HTTP request.
//
// Summary: Determines the client's IP address.
//
// Parameters:
//   - r: *http.Request. The HTTP request to inspect.
//   - trustProxy: bool. If true, trusts 'X-Real-IP' and 'X-Forwarded-For' headers. If false, only uses 'RemoteAddr'.
//
// Returns:
//   - string: The best-effort client IP address.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// RemoteIPFromContext retrieves the remote IP address stored in the context.
//
// IsPrivateNetworkIP checks if the IP address belongs to a private network.
//
// Summary: Checks if an IP is a private network address.
//
// IsPrivateIP checks if the IP address is private, link-local, or loopback.
//
// Summary: Checks if an IP is internal/private.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// This is a comprehensive check for any "internal" IP address that shouldn't be publicly routable.
//
// Parameters:
//   - ip: net.IP. The IP address to check.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - bool: True if the IP is private, link-local, or loopback.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func IsPrivateIP(ip net.IP) bool {
	return validation.IsPrivateIP(ip)
}

func isNAT64Loopback(ip net.IP) bool {
	return validation.IsNAT64Loopback(ip)
}

func isNAT64LinkLocal(ip net.IP) bool {
	return validation.IsNAT64LinkLocal(ip)
}
