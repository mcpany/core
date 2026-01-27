// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"
	"time"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/discovery"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistry is a manual mock for ServiceRegistryInterface
type MockServiceRegistry struct {
	serviceregistry.ServiceRegistryInterface
	services []*configv1.UpstreamServiceConfig
	errors   map[string]string
}

func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	return m.services, nil
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	for _, s := range m.services {
		if s.GetId() == serviceID {
			return s, true
		}
	}
	return nil, false
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	err, ok := m.errors[serviceID]
	return err, ok
}

// MockDiscoveryProvider is a manual mock for discovery.Provider
type MockDiscoveryProvider struct {
	name     string
	services []*configv1.UpstreamServiceConfig
	err      error
}

func (m *MockDiscoveryProvider) Name() string {
	return m.name
}

func (m *MockDiscoveryProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return m.services, m.err
}

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	sr := &MockServiceRegistry{}

	s := NewServer(nil, tm, sr, store, nil, nil)
	assert.NotNil(t, s)
}

func TestServer_UserManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	sr := &MockServiceRegistry{}
	s := NewServer(nil, tm, sr, store, nil, nil)
	ctx := context.Background()

	// Test CreateUser
	basicAuth := &configv1.BasicAuth{}
	basicAuth.SetPasswordHash("plaintext")

	auth := &configv1.Authentication{}
	auth.SetBasicAuth(basicAuth)

	user := &configv1.User{}
	user.SetId("user1")
	user.SetAuthentication(auth)
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
	svc1 := &configv1.UpstreamServiceConfig{}
	svc1.SetName("svc1")
	svc1.SetId("svc1")

	svcError := &configv1.UpstreamServiceConfig{}
	svcError.SetName("svc_error")
	svcError.SetId("svc_error")

	sr := &MockServiceRegistry{
		services: []*configv1.UpstreamServiceConfig{svc1, svcError},
		errors: map[string]string{
			"svc_error": "failed to start",
		},
	}
	s := NewServer(nil, tm, sr, store, nil, nil)
	ctx := context.Background()

	// Test ListServices
	listResp, err := s.ListServices(ctx, &pb.ListServicesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Services, 2)
	assert.Len(t, listResp.ServiceStates, 2)

	// Check svc1 (Healthy)
	svc1State := listResp.ServiceStates[0]
	assert.Equal(t, "svc1", svc1State.Config.GetName())
	assert.Equal(t, "OK", svc1State.GetStatus())

	// Check svc_error (Error)
	svcErrorState := listResp.ServiceStates[1]
	assert.Equal(t, "svc_error", svcErrorState.Config.GetName())
	assert.Equal(t, "ERROR", svcErrorState.GetStatus())
	assert.Equal(t, "failed to start", svcErrorState.GetError())

	// Test GetService (Healthy)
	getResp, err := s.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("svc1")})
	require.NoError(t, err)
	assert.Equal(t, "svc1", getResp.Service.GetName())
	assert.Equal(t, "OK", getResp.ServiceState.GetStatus())

	// Test GetService (Error)
	getResp, err = s.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("svc_error")})
	require.NoError(t, err)
	assert.Equal(t, "svc_error", getResp.Service.GetName())
	assert.Equal(t, "ERROR", getResp.ServiceState.GetStatus())
	assert.Equal(t, "failed to start", getResp.ServiceState.GetError())

	// Test GetService Not Found
	_, err = s.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("unknown")})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestServer_ToolManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	sr := &MockServiceRegistry{}
	s := NewServer(nil, tm, sr, store, nil, nil)
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
	s := NewServer(nil, nil, nil, nil, nil, nil)
	_, err := s.ClearCache(context.Background(), &pb.ClearCacheRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Test ClearCache with valid cache
	realMiddleware := middleware.NewCachingMiddleware(tool.NewMockManagerInterface(ctrl))
	sValid := NewServer(realMiddleware, nil, nil, nil, nil, nil)
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
	sr := &MockServiceRegistry{}
	s := NewServer(nil, tm, sr, ms, nil, nil)
	ctx := context.Background()

	errInternal := status.Error(codes.Internal, "storage error")

	u1 := &configv1.User{}
	u1.SetId("user1")

	// CreateUser Error
	ms.createUserErr = errInternal
	_, err := s.CreateUser(ctx, &pb.CreateUserRequest{User: u1})
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
	_, err = s.UpdateUser(ctx, &pb.UpdateUserRequest{User: u1})
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
	sr := &MockServiceRegistry{}
	s := NewServer(nil, tm, sr, store, nil, nil)
	ctx := context.Background()

	longPassword := string(make([]byte, 73)) // 73 bytes > 72 bytes limit for bcrypt
	basicAuth := &configv1.BasicAuth{}
	basicAuth.SetPasswordHash(longPassword)

	auth := &configv1.Authentication{}
	auth.SetBasicAuth(basicAuth)

	user := &configv1.User{}
	user.SetId("user1")
	user.SetAuthentication(auth)

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
	ctx := context.Background()

	// Fallback to toolManager if serviceRegistry doesn't find it?
	// But in our implementation if serviceRegistry is set, we ONLY use serviceRegistry.
	// So let's test with nil serviceRegistry to trigger fallback path, which tests old logic.

	sFallback := NewServer(nil, tm, nil, store, nil, nil)

	// GetService: Service found but config nil (via ToolManager)
	tm.EXPECT().GetServiceInfo("svc_no_config").Return(&tool.ServiceInfo{Config: nil}, true)
	_, err := sFallback.GetService(ctx, &pb.GetServiceRequest{ServiceId: proto.String("svc_no_config")})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestServer_GetDiscoveryStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	sr := &MockServiceRegistry{}

	// Create manager
	dm := discovery.NewManager()

	// Register provider
	svcDisc := &configv1.UpstreamServiceConfig{}
	svcDisc.SetName("discovered-service")

	provider := &MockDiscoveryProvider{
		name: "test-provider",
		services: []*configv1.UpstreamServiceConfig{svcDisc},
	}
	dm.RegisterProvider(provider)

	// Run discovery
	ctx := context.Background()
	dm.Run(ctx)

	s := NewServer(nil, tm, sr, store, dm, nil)

	// Test GetDiscoveryStatus
	resp, err := s.GetDiscoveryStatus(ctx, &pb.GetDiscoveryStatusRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Providers, 1)
	assert.Equal(t, "test-provider", resp.Providers[0].GetName())
	assert.Equal(t, "OK", resp.Providers[0].GetStatus())
	assert.Equal(t, int32(1), resp.Providers[0].GetDiscoveredCount())

	// Test with nil manager
	sNil := NewServer(nil, tm, sr, store, nil, nil)
	respNil, err := sNil.GetDiscoveryStatus(ctx, &pb.GetDiscoveryStatusRequest{})
	require.NoError(t, err)
	assert.Empty(t, respNil.Providers)
}

// MockAuditStore is a manual mock for audit.Store
type MockAuditStore struct {
	entries []audit.Entry
	readErr error
	writeErr error
	closeErr error
}

func (m *MockAuditStore) Write(ctx context.Context, entry audit.Entry) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	// For simplicity, just return all entries
	return m.entries, nil
}

func (m *MockAuditStore) Close() error {
	return m.closeErr
}

func TestServer_ListAuditLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	sr := &MockServiceRegistry{}

	// Create audit middleware
	auditCfg := &configv1.AuditConfig{}
	auditCfg.SetEnabled(true)
	am, err := middleware.NewAuditMiddleware(auditCfg)
	require.NoError(t, err)

	// Inject mock store
	now := time.Now()
	mockStore := &MockAuditStore{
		entries: []audit.Entry{
			{
				Timestamp: now,
				ToolName: "test-tool",
				UserID: "user1",
				Duration: "100ms",
				DurationMs: 100,
			},
		},
	}
	am.SetStore(mockStore)

	s := NewServer(nil, tm, sr, store, nil, am)
	ctx := context.Background()

	// Test ListAuditLogs - Success
	resp, err := s.ListAuditLogs(ctx, &pb.ListAuditLogsRequest{
		StartTime: proto.String(now.Add(-1 * time.Hour).Format(time.RFC3339)),
		EndTime: proto.String(now.Add(1 * time.Hour).Format(time.RFC3339)),
	})
	require.NoError(t, err)
	require.Len(t, resp.Entries, 1)
	assert.Equal(t, "test-tool", resp.Entries[0].GetToolName())
	assert.Equal(t, "user1", resp.Entries[0].GetUserId())

	// Test ListAuditLogs - Audit disabled (middleware nil)
	sNil := NewServer(nil, tm, sr, store, nil, nil)
	_, err = sNil.ListAuditLogs(ctx, &pb.ListAuditLogsRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Test ListAuditLogs - Read Error
	mockStore.readErr = assert.AnError
	_, err = s.ListAuditLogs(ctx, &pb.ListAuditLogsRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))

	// Test ListAuditLogs - Invalid Time Format
	_, err = s.ListAuditLogs(ctx, &pb.ListAuditLogsRequest{StartTime: proto.String("invalid")})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = s.ListAuditLogs(ctx, &pb.ListAuditLogsRequest{EndTime: proto.String("invalid")})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestServer_ListServices_Fallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()

	s := NewServer(nil, tm, nil, store, nil, nil)
	ctx := context.Background()

	svcFallback := &configv1.UpstreamServiceConfig{}
	svcFallback.SetId("svc_fallback")
	svcFallback.SetName("svc_fallback")

	// Mock ToolManager returning services
	tm.EXPECT().ListServices().Return([]*tool.ServiceInfo{
		{
			Config: svcFallback,
		},
		{
			Config: nil, // Should be ignored
		},
	})

	resp, err := s.ListServices(ctx, &pb.ListServicesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Services, 1)
	assert.Equal(t, "svc_fallback", resp.Services[0].GetName())
	assert.Equal(t, "OK", resp.ServiceStates[0].GetStatus())
}
