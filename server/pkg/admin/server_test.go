// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)

	s := NewServer(nil, tm, store)
	assert.NotNil(t, s)
}

func TestServer_UserManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	s := NewServer(nil, tm, store)
	ctx := context.Background()

	// Test CreateUser
	user := &configv1.User{
		Id: proto.String("user1"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					PasswordHash: proto.String("plaintext"),
				},
			},
		},
	}
	createResp, err := s.CreateUser(ctx, &pb.CreateUserRequest{User: user})
	require.NoError(t, err)
	assert.Equal(t, "user1", createResp.User.GetId())
	assert.NotEqual(t, "plaintext", createResp.User.Authentication.GetBasicAuth().GetPasswordHash())
	assert.Contains(t, createResp.User.Authentication.GetBasicAuth().GetPasswordHash(), "$2")

	// Test CreateUser validation
	_, err = s.CreateUser(ctx, &pb.CreateUserRequest{User: nil})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Test GetUser
	getResp, err := s.GetUser(ctx, &pb.GetUserRequest{UserId: proto.String("user1")})
	require.NoError(t, err)
	assert.Equal(t, "user1", getResp.User.GetId())

	_, err = s.GetUser(ctx, &pb.GetUserRequest{UserId: proto.String("nonexistent")})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	// Test ListUsers
	listResp, err := s.ListUsers(ctx, &pb.ListUsersRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Users, 1)

	// Test UpdateUser
	user.Roles = []string{"admin"}
	updateResp, err := s.UpdateUser(ctx, &pb.UpdateUserRequest{User: user})
	require.NoError(t, err)
	assert.Equal(t, []string{"admin"}, updateResp.User.Roles)

	_, err = s.UpdateUser(ctx, &pb.UpdateUserRequest{User: nil})
	assert.Error(t, err)

	// Test DeleteUser
	deleteResp, err := s.DeleteUser(ctx, &pb.DeleteUserRequest{UserId: proto.String("user1")})
	require.NoError(t, err)
	assert.NotNil(t, deleteResp)

	_, err = s.GetUser(ctx, &pb.GetUserRequest{UserId: proto.String("user1")})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestServer_ServiceManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	s := NewServer(nil, tm, store)
	ctx := context.Background()

	// Test ListServices
	expectedServices := []*tool.ServiceInfo{
		{
			Name: "svc1",
			Config: &configv1.UpstreamServiceConfig{
				Name: proto.String("svc1"),
			},
		},
	}
	tm.EXPECT().ListServices().Return(expectedServices)

	listResp, err := s.ListServices(ctx, &pb.ListServicesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Services, 1)
	assert.Equal(t, "svc1", listResp.Services[0].GetName())

	// Test GetService
	tm.EXPECT().GetServiceInfo("svc1").Return(expectedServices[0], true)
	getResp, err := s.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("svc1")})
	require.NoError(t, err)
	assert.Equal(t, "svc1", getResp.Service.GetName())

	// Test GetService Not Found
	tm.EXPECT().GetServiceInfo("unknown").Return(nil, false)
	_, err = s.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("unknown")})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestServer_ToolManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	s := NewServer(nil, tm, store)
	ctx := context.Background()

	// Mock Tool
	mockTool := &tool.MockTool{
		ToolFunc: func() *mcprouterv1.Tool {
			return &mcprouterv1.Tool{Name: proto.String("tool1")}
		},
	}

	// Test ListTools
	tm.EXPECT().ListTools().Return([]tool.Tool{mockTool})
	listResp, err := s.ListTools(ctx, &pb.ListToolsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Tools, 1)
	assert.Equal(t, "tool1", listResp.Tools[0].GetName())

	// Test GetTool
	tm.EXPECT().GetTool("tool1").Return(mockTool, true)
	getResp, err := s.GetTool(ctx, &pb.GetToolRequest{ToolName: proto.String("tool1")})
	require.NoError(t, err)
	assert.Equal(t, "tool1", getResp.Tool.GetName())

	// Test GetTool Not Found
	tm.EXPECT().GetTool("unknown").Return(nil, false)
	_, err = s.GetTool(ctx, &pb.GetToolRequest{ToolName: proto.String("unknown")})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestServer_ClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test with nil cache
	s := NewServer(nil, nil, nil)
	_, err := s.ClearCache(context.Background(), &pb.ClearCacheRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Test ClearCache with valid cache
	realMiddleware := middleware.NewCachingMiddleware(tool.NewMockManagerInterface(ctrl))
	sValid := NewServer(realMiddleware, nil, nil)
	resp, err := sValid.ClearCache(context.Background(), &pb.ClearCacheRequest{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// mockStorage is a mock implementation of storage.Storage for testing errors
type mockStorage struct {
	*memory.Store
	createUserErr error
	getUserErr    error
	listUsersErr  error
	updateUserErr error
	deleteUserErr error
}

func (m *mockStorage) CreateUser(ctx context.Context, user *configv1.User) error {
	if m.createUserErr != nil {
		return m.createUserErr
	}
	return m.Store.CreateUser(ctx, user)
}

func (m *mockStorage) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	return m.Store.GetUser(ctx, id)
}

func (m *mockStorage) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	if m.listUsersErr != nil {
		return nil, m.listUsersErr
	}
	return m.Store.ListUsers(ctx)
}

func (m *mockStorage) UpdateUser(ctx context.Context, user *configv1.User) error {
	if m.updateUserErr != nil {
		return m.updateUserErr
	}
	return m.Store.UpdateUser(ctx, user)
}

func (m *mockStorage) DeleteUser(ctx context.Context, id string) error {
	if m.deleteUserErr != nil {
		return m.deleteUserErr
	}
	return m.Store.DeleteUser(ctx, id)
}

func TestServer_UserManagement_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)
	ms := &mockStorage{Store: memory.NewStore()}
	s := NewServer(nil, tm, ms)
	ctx := context.Background()

	errInternal := status.Error(codes.Internal, "storage error")

	// CreateUser Error
	ms.createUserErr = errInternal
	_, err := s.CreateUser(ctx, &pb.CreateUserRequest{User: &configv1.User{Id: proto.String("user1")}})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.createUserErr = nil

	// GetUser Error
	ms.getUserErr = errInternal
	_, err = s.GetUser(ctx, &pb.GetUserRequest{UserId: proto.String("user1")})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.getUserErr = nil

	// ListUsers Error
	ms.listUsersErr = errInternal
	_, err = s.ListUsers(ctx, &pb.ListUsersRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.listUsersErr = nil

	// UpdateUser Error
	ms.updateUserErr = errInternal
	_, err = s.UpdateUser(ctx, &pb.UpdateUserRequest{User: &configv1.User{Id: proto.String("user1")}})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.updateUserErr = nil

	// DeleteUser Error
	ms.deleteUserErr = errInternal
	_, err = s.DeleteUser(ctx, &pb.DeleteUserRequest{UserId: proto.String("user1")})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.deleteUserErr = nil
}

func TestServer_UserManagement_PasswordHashing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	s := NewServer(nil, tm, store)
	ctx := context.Background()

	longPassword := string(make([]byte, 73)) // 73 bytes > 72 bytes limit for bcrypt
	user := &configv1.User{
		Id: proto.String("user1"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					PasswordHash: proto.String(longPassword),
				},
			},
		},
	}

	// CreateUser - Password hashing failure
	_, err := s.CreateUser(ctx, &pb.CreateUserRequest{User: user})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "failed to hash password")

	// UpdateUser - Password hashing failure
	_, err = s.UpdateUser(ctx, &pb.UpdateUserRequest{User: user})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "failed to hash password")
}

func TestServer_ServiceManagement_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	s := NewServer(nil, tm, store)
	ctx := context.Background()

	// GetService: Service found but config nil
	tm.EXPECT().GetServiceInfo("svc_no_config").Return(&tool.ServiceInfo{Config: nil}, true)
	_, err := s.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("svc_no_config")})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
