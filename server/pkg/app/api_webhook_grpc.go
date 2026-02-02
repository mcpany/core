// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	pb "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/webhook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WebhookServiceServer implements the WebhookService gRPC interface.
type WebhookServiceServer struct {
	pb.UnimplementedWebhookServiceServer
	manager *webhook.Manager
}

// NewWebhookServiceServer creates a new WebhookServiceServer.
func NewWebhookServiceServer(manager *webhook.Manager) *WebhookServiceServer {
	return &WebhookServiceServer{
		manager: manager,
	}
}

// ListWebhooks lists all available webhooks.
func (s *WebhookServiceServer) ListWebhooks(_ context.Context, _ *pb.ListWebhooksRequest) (*pb.ListWebhooksResponse, error) {
	webhooks := s.manager.ListWebhooks()
	return pb.ListWebhooksResponse_builder{
		Webhooks: webhooks,
	}.Build(), nil
}

// CreateWebhook creates a new webhook.
func (s *WebhookServiceServer) CreateWebhook(_ context.Context, req *pb.CreateWebhookRequest) (*pb.CreateWebhookResponse, error) {
	if req.GetWebhook() == nil {
		return nil, status.Error(codes.InvalidArgument, "webhook is required")
	}

	created, err := s.manager.CreateWebhook(req.GetWebhook())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create webhook: %v", err)
	}

	return pb.CreateWebhookResponse_builder{
		Webhook: created,
	}.Build(), nil
}

// DeleteWebhook deletes a webhook.
func (s *WebhookServiceServer) DeleteWebhook(_ context.Context, req *pb.DeleteWebhookRequest) (*pb.DeleteWebhookResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.manager.DeleteWebhook(req.GetId()); err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to delete webhook: %v", err)
	}

	return &pb.DeleteWebhookResponse{}, nil
}
