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

// GrpcStatsHandler - Auto-generated documentation.
//
// Summary: GrpcStatsHandler is a gRPC stats handler that records metrics for RPCs and connections.
//
// Fields:
//   - Various fields for GrpcStatsHandler.
type GrpcStatsHandler struct {
	Wrapped stats.Handler
}

// TagRPC can be used to tag RPCs with custom information. Parameters: - ctx: The context of the RPC. - info: Information about the RPC tag. Returns: - The context, potentially modified with new tags.
//
// Summary: TagRPC can be used to tag RPCs with custom information. Parameters: - ctx: The context of the RPC. - info: Information about the RPC tag. Returns: - The context, potentially modified with new tags.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - info (*stats.RPCTagInfo): The info parameter used in the operation.
//
// Returns:
//   - (context.Context): The resulting context.Context object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagRPC(ctx, info)
	}
	return ctx
}

// HandleRPC - Auto-generated documentation.
//
// Summary: HandleRPC processes RPC stats and increments counters for started and finished RPCs.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// TagConn can be used to tag connections with custom information. Parameters: - ctx: The context of the connection. - info: Information about the connection tag. Returns: - The context, potentially modified with new tags.
//
// Summary: TagConn can be used to tag connections with custom information. Parameters: - ctx: The context of the connection. - info: Information about the connection tag. Returns: - The context, potentially modified with new tags.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - info (*stats.ConnTagInfo): The info parameter used in the operation.
//
// Returns:
//   - (context.Context): The resulting context.Context object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagConn(ctx, info)
	}
	return ctx
}

// HandleConn - Auto-generated documentation.
//
// Summary: HandleConn processes connection stats and increments counters for opened and closed connections.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
