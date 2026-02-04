// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package dashboard implements the dashboard service.
package dashboard

import (
	"context"

	pb "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Server implements the DashboardServiceServer interface.
type Server struct {
	pb.UnimplementedDashboardServiceServer
	storage storage.Storage
}

// NewServer creates a new DashboardServer.
func NewServer(storage storage.Storage) *Server {
	return &Server{storage: storage}
}

// GetDashboardLayout retrieves the dashboard layout for the authenticated user.
func (s *Server) GetDashboardLayout(ctx context.Context, _ *pb.GetDashboardLayoutRequest) (*pb.GetDashboardLayoutResponse, error) {
	// Extract user ID from context
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		// Fallback for dev mode / no auth - default to "system-admin" or empty
		userID = "system-admin"
	}

	user, err := s.storage.GetUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	if user == nil || user.GetPreferences() == nil {
		return &pb.GetDashboardLayoutResponse{}, nil
	}

	return &pb.GetDashboardLayoutResponse{
		LayoutJson: user.GetPreferences().GetDashboardLayoutJson(),
	}, nil
}

// SaveDashboardLayout saves the dashboard layout for the authenticated user.
func (s *Server) SaveDashboardLayout(ctx context.Context, req *pb.SaveDashboardLayoutRequest) (*pb.SaveDashboardLayoutResponse, error) {
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		userID = "system-admin"
	}

	user, err := s.storage.GetUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	if user == nil {
		user = configv1.User_builder{
			Id: proto.String(userID),
		}.Build()
	}

	// Create a mutable copy if needed, but protobuf objects are mutable pointers in Go
	// However, we should be careful if it came from a cache or read-only source.
	// storage.GetUser usually unmarshals a fresh copy.

	// Ensure Preferences struct exists
	if user.GetPreferences() == nil {
		user.SetPreferences(configv1.UserPreferences_builder{}.Build())
	}
	user.GetPreferences().SetDashboardLayoutJson(req.GetLayoutJson())

	// Upsert logic
	if err := s.storage.UpdateUser(ctx, user); err != nil {
		// If update fails, try create
		if err := s.storage.CreateUser(ctx, user); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to save user preferences: %v", err)
		}
	}

	return &pb.SaveDashboardLayoutResponse{}, nil
}
