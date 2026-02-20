// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"fmt"
	"time"

	pb "github.com/mcpany/core/proto/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ===================================================================
// Webhook Management
// ===================================================================

// CreateSystemWebhook creates a new system webhook.
func (s *Server) CreateSystemWebhook(ctx context.Context, req *pb.CreateSystemWebhookRequest) (*pb.CreateSystemWebhookResponse, error) {
	if !req.HasWebhook() {
		return nil, status.Error(codes.InvalidArgument, "webhook is required")
	}
	webhook := req.GetWebhook()
	if webhook.GetId() == "" {
		// Auto-generate ID if missing
		webhook.SetId(fmt.Sprintf("wh-%d", time.Now().UnixNano()))
	}
	if webhook.GetCreatedAt() == "" {
		webhook.SetCreatedAt(time.Now().Format(time.RFC3339))
	}

	if err := s.storage.CreateSystemWebhook(ctx, webhook); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create webhook: %v", err)
	}

	return pb.CreateSystemWebhookResponse_builder{Webhook: webhook}.Build(), nil
}

// ListSystemWebhooks returns all registered system webhooks.
func (s *Server) ListSystemWebhooks(ctx context.Context, _ *pb.ListSystemWebhooksRequest) (*pb.ListSystemWebhooksResponse, error) {
	webhooks, err := s.storage.ListSystemWebhooks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list webhooks: %v", err)
	}
	return pb.ListSystemWebhooksResponse_builder{Webhooks: webhooks}.Build(), nil
}

// DeleteSystemWebhook deletes a system webhook by ID.
func (s *Server) DeleteSystemWebhook(ctx context.Context, req *pb.DeleteSystemWebhookRequest) (*pb.DeleteSystemWebhookResponse, error) {
	if err := s.storage.DeleteSystemWebhook(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete webhook: %v", err)
	}
	return &pb.DeleteSystemWebhookResponse{}, nil
}

// UpdateSystemWebhook updates an existing system webhook.
func (s *Server) UpdateSystemWebhook(ctx context.Context, req *pb.UpdateSystemWebhookRequest) (*pb.UpdateSystemWebhookResponse, error) {
	if !req.HasWebhook() {
		return nil, status.Error(codes.InvalidArgument, "webhook is required")
	}
	if err := s.storage.UpdateSystemWebhook(ctx, req.GetWebhook()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update webhook: %v", err)
	}
	return pb.UpdateSystemWebhookResponse_builder{Webhook: req.GetWebhook()}.Build(), nil
}

// TestSystemWebhook tests a system webhook delivery.
func (s *Server) TestSystemWebhook(ctx context.Context, req *pb.TestSystemWebhookRequest) (*pb.TestSystemWebhookResponse, error) {
	webhook, err := s.storage.GetSystemWebhook(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get webhook: %v", err)
	}
	if webhook == nil {
		return nil, status.Error(codes.NotFound, "webhook not found")
	}

	// TODO: Implement actual test dispatch via Dispatcher.
	// For now, returning simulated success.
	return pb.TestSystemWebhookResponse_builder{
		Success:    proto.Bool(true),
		Message:    proto.String("Test payload sent successfully (Simulated)"),
		StatusCode: proto.Int32(200),
	}.Build(), nil
}
