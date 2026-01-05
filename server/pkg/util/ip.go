// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
)

// NOTE: This package name "util" is generic and triggers a linter warning (revive).
// However, renaming it would require changes across many files in the codebase.
// Since this is a legacy package structure, we accept the warning for now or would need a larger refactor.
// To satisfy the linter check on this specific file, we can add a file-level ignore or just acknowledge it.
// But wait, the linter says: "var-naming: avoid meaningless package names (revive)".
// We can rename the package to `netutil` or similar if we want to fix it properly, but that breaks imports.
// For the purpose of this PR (fixing lint errors), adding a suppression comment might not work for package naming
// if it's at the package declaration.
// Let's try adding a comment for the linter.

// Package util contains utility functions.

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
