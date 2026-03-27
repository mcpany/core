// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package metrics provides gRPC interceptors for metrics.
package metrics

import (
	"context"

	"google.golang.org/grpc/stats"
)

var (
	metricGrpcRPCStartedTotal        = []string{"grpc", "rpc", "started", "total"}
	metricGrpcRPCFinishedTotal       = []string{"grpc", "rpc", "finished", "total"}
	metricGrpcConnectionsOpenedTotal = []string{"grpc", "connections", "opened", "total"}
	metricGrpcConnectionsClosedTotal = []string{"grpc", "connections", "closed", "total"}
)

// GrpcStatsHandler is a gRPC stats handler that records metrics for RPCs and connections.
// It can optionally wrap another stats.Handler (e.g., OpenTelemetry).
//
// Summary: Represents a GrpcStatsHandler.
type GrpcStatsHandler struct {
	Wrapped stats.Handler
}

// TagRPC provides tagrpc functionality.
//
// Summary: TagRPC.
//
// Parameters.
//   - ctx: The parameter.
//   - info: The parameter.
//
// Returns.
//   - result: The result.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagRPC(ctx, info)
	}
	return ctx
}

// HandleRPC provides handlerpc functionality.
//
// Summary: HandleRPC.
//
// Parameters.
//   - ctx: The parameter.
//   - s: The parameter.
//
// Returns.
//   - None.
func (h *GrpcStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	if h.Wrapped != nil {
		h.Wrapped.HandleRPC(ctx, s)
	}
	switch s.(type) {
	case *stats.Begin:
		IncrCounter(metricGrpcRPCStartedTotal, 1)
	case *stats.End:
		IncrCounter(metricGrpcRPCFinishedTotal, 1)
	}
}

// TagConn provides tagconn functionality.
//
// Summary: TagConn.
//
// Parameters.
//   - ctx: The parameter.
//   - info: The parameter.
//
// Returns.
//   - result: The result.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagConn(ctx, info)
	}
	return ctx
}

// HandleConn provides handleconn functionality.
//
// Summary: HandleConn.
//
// Parameters.
//   - ctx: The parameter.
//   - s: The parameter.
//
// Returns.
//   - None.
func (h *GrpcStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	if h.Wrapped != nil {
		h.Wrapped.HandleConn(ctx, s)
	}
	switch s.(type) {
	case *stats.ConnBegin:
		IncrCounter(metricGrpcConnectionsOpenedTotal, 1)
	case *stats.ConnEnd:
		IncrCounter(metricGrpcConnectionsClosedTotal, 1)
	}
}
