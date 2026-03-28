package app

import (
	"context"
	"google.golang.org/grpc"
)

func (a *Application) registerSkillGRPCService(grpcServer *grpc.Server) {}
func (a *Application) registerSkillGRPCGateway(ctx context.Context, gwmux interface{}, endpoint string, opts []grpc.DialOption) error {
	return nil
}
