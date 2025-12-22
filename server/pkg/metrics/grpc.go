// Package metrics provides gRPC interceptors for metrics.

package metrics

import (
	"context"

	"google.golang.org/grpc/stats"
)

// GrpcStatsHandler is a gRPC stats handler that records metrics for RPCs and connections.
type GrpcStatsHandler struct{}

// TagRPC can be used to tag RPCs with custom information.
//
// Parameters:
//   - ctx: The context of the RPC.
//   - _ : Information about the RPC tag (unused).
//
// Returns:
//   - The context, potentially modified with new tags.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC processes RPC stats and increments counters for started and finished RPCs.
//
// Parameters:
//   - ctx: The context of the RPC.
//   - s: The RPC stats.
func (h *GrpcStatsHandler) HandleRPC(_ context.Context, s stats.RPCStats) {
	switch s.(type) {
	case *stats.Begin:
		IncrCounter([]string{"grpc", "rpc", "started", "total"}, 1)
	case *stats.End:
		IncrCounter([]string{"grpc", "rpc", "finished", "total"}, 1)
	}
}

// TagConn can be used to tag connections with custom information.
//
// Parameters:
//   - ctx: The context of the connection.
//   - _ : Information about the connection tag (unused).
//
// Returns:
//   - The context, potentially modified with new tags.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes connection stats and increments counters for opened and closed connections.
//
// Parameters:
//   - ctx: The context of the connection.
//   - s: The connection stats.
func (h *GrpcStatsHandler) HandleConn(_ context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
		IncrCounter([]string{"grpc", "connections", "opened", "total"}, 1)
	case *stats.ConnEnd:
		IncrCounter([]string{"grpc", "connections", "closed", "total"}, 1)
	}
}
