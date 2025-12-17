// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net"
	"strings"

	"github.com/mcpany/core/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// IPAllowlistInterceptor creates gRPC interceptors (Unary and Stream) that restrict access
// to a list of allowed IP addresses or CIDR blocks.
type IPAllowlistInterceptor struct {
	ipNets []*net.IPNet
}

// NewIPAllowlistInterceptor creates a new IPAllowlistInterceptor.
func NewIPAllowlistInterceptor(allowedIPs []string) *IPAllowlistInterceptor {
	if len(allowedIPs) == 0 {
		return &IPAllowlistInterceptor{}
	}

	var ipNets []*net.IPNet
	for _, ipStr := range allowedIPs {
		if !strings.Contains(ipStr, "/") {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				logging.GetLogger().Warn("Invalid IP address in allowlist", "ip", ipStr)
				continue
			}
			if ip.To4() != nil {
				ipStr += "/32"
			} else {
				ipStr += "/128"
			}
		}
		_, ipNet, err := net.ParseCIDR(ipStr)
		if err != nil {
			logging.GetLogger().Warn("Invalid CIDR in allowlist", "cidr", ipStr, "error", err)
			continue
		}
		ipNets = append(ipNets, ipNet)
	}
	return &IPAllowlistInterceptor{ipNets: ipNets}
}

func (i *IPAllowlistInterceptor) checkIP(ctx context.Context) error {
	if len(i.ipNets) == 0 {
		return nil
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logging.GetLogger().Warn("Could not get peer from context")
		return status.Error(codes.Unauthenticated, "no peer information")
	}

	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	// Handle [IPv6]
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")

	ip := net.ParseIP(host)
	if ip == nil {
		logging.GetLogger().Warn("Could not parse peer IP", "addr", addr)
		return status.Error(codes.PermissionDenied, "invalid ip")
	}

	for _, ipNet := range i.ipNets {
		if ipNet.Contains(ip) {
			return nil
		}
	}

	logging.GetLogger().Warn("Access denied (gRPC)", "remote_ip", host)
	return status.Error(codes.PermissionDenied, "access denied")
}

// UnaryServerInterceptor returns a UnaryServerInterceptor.
func (i *IPAllowlistInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := i.checkIP(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a StreamServerInterceptor.
func (i *IPAllowlistInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := i.checkIP(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}
