// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package metrics provides gRPC interceptors for metrics.
// Summary: GrpcStatsHandler is a gRPC stats handler that records metrics for RPCs and connections.
// It can optionally wrap another stats.Handler (e.g., OpenTelemetry).
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// TagRPC can be used to tag RPCs with custom information.
//
// Parameters:
//   - ctx: The context of the RPC.
//   - info: Information about the RPC tag.
//
// Returns:
//   - The context, potentially modified with new tags.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// HandleRPC processes RPC stats and increments counters for started and finished RPCs.
//
// Parameters:
//   - ctx: The context of the RPC.
//   - s: The RPC stats.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// TagConn can be used to tag connections with custom information.
//
// Parameters:
//   - ctx: The context of the connection.
//   - info: Information about the connection tag.
//
// Returns:
//   - The context, potentially modified with new tags.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// HandleConn processes connection stats and increments counters for opened and closed connections.
//
// Parameters:
//   - ctx: The context of the connection.
//   - s: The connection stats.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
package metrics

import (
	"context"

	"google.golang.org/grpc/stats"
)

var (
	metricGrpcRPCStartedTotal		= []string{"grpc", "rpc", "started", "total"}
	metricGrpcRPCFinishedTotal		= []string{"grpc", "rpc", "finished", "total"}
	metricGrpcConnectionsOpenedTotal	= []string{"grpc", "connections", "opened", "total"}
	metricGrpcConnectionsClosedTotal	= []string{"grpc", "connections", "closed", "total"}
)

type GrpcStatsHandler struct {
	Wrapped stats.Handler
}

func (h *GrpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagRPC(ctx, info)
	}
	return ctx
}

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

func (h *GrpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if h.Wrapped != nil {
		ctx = h.Wrapped.TagConn(ctx, info)
	}
	return ctx
}

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
