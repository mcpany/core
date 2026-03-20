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
// Summary: GrpcStatsHandler is a gRPC stats handler that records metrics for RPCs and connections.
type GrpcStatsHandler struct {
	Wrapped stats.Handler
}

// TagRPC can be used to tag RPCs with custom information.
//
// Summary: TagRPC can be used to tag RPCs with custom information.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - info (*stats.RPCTagInfo): The provided info data.
//
// Returns:
//   - context.Context: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagRPC(ctx, info)
	}
	return ctx
}

// HandleRPC processes RPC stats and increments counters for started and finished RPCs.
//
// Summary: HandleRPC processes RPC stats and increments counters for started and finished RPCs.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - s (stats.RPCStats): The provided s data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Summary: TagConn can be used to tag connections with custom information.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - info (*stats.ConnTagInfo): The provided info data.
//
// Returns:
//   - context.Context: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagConn(ctx, info)
	}
	return ctx
}

// HandleConn processes connection stats and increments counters for opened and closed connections.
//
// Summary: HandleConn processes connection stats and increments counters for opened and closed connections.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - s (stats.ConnStats): The provided s data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
