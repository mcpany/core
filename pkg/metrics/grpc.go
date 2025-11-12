
package metrics

import (
	"context"

	"google.golang.org/grpc/stats"
)

// GrpcStatsHandler is a gRPC stats handler that records metrics.
type GrpcStatsHandler struct{}

// TagRPC can be used to tag RPCs with custom information.
func (h *GrpcStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC processes RPC stats.
func (h *GrpcStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {}

// TagConn can be used to tag connections with custom information.
func (h *GrpcStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes connection stats.
func (h *GrpcStatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
		IncrCounter([]string{"grpc", "connections", "opened", "total"}, 1)
	case *stats.ConnEnd:
		IncrCounter([]string{"grpc", "connections", "closed", "total"}, 1)
	}
}
