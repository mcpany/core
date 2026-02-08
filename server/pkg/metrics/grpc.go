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
//
// Summary: is a gRPC stats handler that records metrics for RPCs and connections.
type GrpcStatsHandler struct {
	Wrapped stats.Handler
}

// TagRPC can be used to tag RPCs with custom information.
//
// Summary: can be used to tag RPCs with custom information.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - info: *stats.RPCTagInfo. The info.
//
// Returns:
//   - context.Context: The context.Context.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagRPC(ctx, info)
	}
	return ctx
}

// HandleRPC processes RPC stats and increments counters for started and finished RPCs.
//
// Summary: processes RPC stats and increments counters for started and finished RPCs.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - s: stats.RPCStats. The s.
//
// Returns:
//   None.
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

// TagConn can be used to tag connections with custom information.
//
// Summary: can be used to tag connections with custom information.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - info: *stats.ConnTagInfo. The info.
//
// Returns:
//   - context.Context: The context.Context.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagConn(ctx, info)
	}
	return ctx
}

// HandleConn processes connection stats and increments counters for opened and closed connections.
//
// Summary: processes connection stats and increments counters for opened and closed connections.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - s: stats.ConnStats. The s.
//
// Returns:
//   None.
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
