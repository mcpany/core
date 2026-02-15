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

	user := configv1.User_builder{
		Id: proto.String("user1"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username:     proto.String("user1"),
				PasswordHash: proto.String("plaintext"),
			}.Build(),
		}.Build(),
	}.Build()
	createResp, err := s.CreateUser(ctx, pb.CreateUserRequest_builder{User: user}.Build())
	require.NoError(t, err)
	assert.Equal(t, "user1", createResp.GetUser().GetId())
	assert.NotEqual(t, "plaintext", createResp.GetUser().GetAuthentication().GetBasicAuth().GetPasswordHash())
	assert.Contains(t, createResp.GetUser().GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2")

	// Test CreateUser validation
	_, err = s.CreateUser(ctx, pb.CreateUserRequest_builder{User: nil}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Test GetUser
	getResp, err := s.GetUser(ctx, pb.GetUserRequest_builder{UserId: proto.String("user1")}.Build())
	require.NoError(t, err)
	assert.Equal(t, "user1", getResp.GetUser().GetId())

	_, err = s.GetUser(ctx, pb.GetUserRequest_builder{UserId: proto.String("nonexistent")}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	// Test ListUsers
	listResp, err := s.ListUsers(ctx, &pb.ListUsersRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.GetUsers(), 1)

	// Test UpdateUser
	user.SetRoles([]string{"admin"})
	updateResp, err := s.UpdateUser(ctx, pb.UpdateUserRequest_builder{User: user}.Build())
	require.NoError(t, err)
	assert.Equal(t, []string{"admin"}, updateResp.GetUser().GetRoles())

	_, err = s.UpdateUser(ctx, pb.UpdateUserRequest_builder{User: nil}.Build())
	assert.Error(t, err)

	// Test DeleteUser
	deleteResp, err := s.DeleteUser(ctx, pb.DeleteUserRequest_builder{UserId: proto.String("user1")}.Build())
	require.NoError(t, err)
	assert.NotNil(t, deleteResp)

	_, err = s.GetUser(ctx, pb.GetUserRequest_builder{UserId: proto.String("user1")}.Build())
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
	assert.Len(t, listResp.GetServices(), 2)
	assert.Len(t, listResp.GetServiceStates(), 2)

	// Check svc1 (Healthy)
	svc1State := listResp.GetServiceStates()[0]
	assert.Equal(t, "svc1", svc1State.GetConfig().GetName())
	assert.Equal(t, "OK", svc1State.GetStatus())

	// Check svc_error (Error)
	svcErrorState := listResp.GetServiceStates()[1]
	assert.Equal(t, "svc_error", svcErrorState.GetConfig().GetName())
	assert.Equal(t, "ERROR", svcErrorState.GetStatus())
	assert.Equal(t, "failed to start", svcErrorState.GetError())

	// Test GetService (Healthy)
	getResp, err := s.GetService(ctx, pb.GetServiceRequest_builder{ServiceId: proto.String("svc1")}.Build())
	require.NoError(t, err)
	assert.Equal(t, "svc1", getResp.GetService().GetName())
	assert.Equal(t, "OK", getResp.GetServiceState().GetStatus())

	// Test GetService (Error)
	getResp, err = s.GetService(ctx, pb.GetServiceRequest_builder{ServiceId: proto.String("svc_error")}.Build())
	require.NoError(t, err)
	assert.Equal(t, "svc_error", getResp.GetService().GetName())
	assert.Equal(t, "ERROR", getResp.GetServiceState().GetStatus())
	assert.Equal(t, "failed to start", getResp.GetServiceState().GetError())

	// Test GetService Not Found
	_, err = s.GetService(ctx, pb.GetServiceRequest_builder{ServiceId: proto.String("unknown")}.Build())
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
			return mcprouterv1.Tool_builder{Name: proto.String("tool1")}.Build()
		},
	}

	// Test ListTools
	tm.EXPECT().ListTools().Return([]tool.Tool{mockTool})
	listResp, err := s.ListTools(ctx, &pb.ListToolsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.GetTools(), 1)
	assert.Equal(t, "tool1", listResp.GetTools()[0].GetName())

	// Test GetTool
	tm.EXPECT().GetTool("tool1").Return(mockTool, true)
	getResp, err := s.GetTool(ctx, pb.GetToolRequest_builder{ToolName: proto.String("tool1")}.Build())
	require.NoError(t, err)
	assert.Equal(t, "tool1", getResp.GetTool().GetName())

	// Test GetTool Not Found
	tm.EXPECT().GetTool("unknown").Return(nil, false)
	_, err = s.GetTool(ctx, pb.GetToolRequest_builder{ToolName: proto.String("unknown")}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestServer_ClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test with nil cache
	s := NewServer(nil, nil, nil, nil, nil, nil)
	_, err := s.ClearCache(context.Background(), pb.ClearCacheRequest_builder{}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	// Test ClearCache with valid cache
	realMiddleware := middleware.NewCachingMiddleware(tool.NewMockManagerInterface(ctrl))
	sValid := NewServer(realMiddleware, nil, nil, nil, nil, nil)
	resp, err := sValid.ClearCache(context.Background(), pb.ClearCacheRequest_builder{}.Build())
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
	_, err := s.CreateUser(ctx, pb.CreateUserRequest_builder{User: u1}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.createUserErr = nil

	// GetUser Error
	ms.getUserErr = errInternal
	_, err = s.GetUser(ctx, pb.GetUserRequest_builder{UserId: proto.String("user1")}.Build())
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
	_, err = s.UpdateUser(ctx, pb.UpdateUserRequest_builder{User: u1}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	ms.updateUserErr = nil

	// DeleteUser Error
	ms.deleteUserErr = errInternal
	_, err = s.DeleteUser(ctx, pb.DeleteUserRequest_builder{UserId: proto.String("user1")}.Build())
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
	_, err := s.CreateUser(ctx, pb.CreateUserRequest_builder{User: user}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "failed to hash password")

	// UpdateUser - Password hashing failure
	_, err = s.UpdateUser(ctx, pb.UpdateUserRequest_builder{User: user}.Build())
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
	_, err := sFallback.GetService(ctx, pb.GetServiceRequest_builder{ServiceId: proto.String("svc_no_config")}.Build())
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
	resp, err := s.GetDiscoveryStatus(ctx, pb.GetDiscoveryStatusRequest_builder{}.Build())
	require.NoError(t, err)
	require.Len(t, resp.GetProviders(), 1)
	assert.Equal(t, "test-provider", resp.GetProviders()[0].GetName())
	assert.Equal(t, "OK", resp.GetProviders()[0].GetStatus())
	assert.Equal(t, int32(1), resp.GetProviders()[0].GetDiscoveredCount())

	// Test with nil manager
	sNil := NewServer(nil, tm, sr, store, nil, nil)
	respNil, err := sNil.GetDiscoveryStatus(ctx, pb.GetDiscoveryStatusRequest_builder{}.Build())
	require.NoError(t, err)
	assert.Empty(t, respNil.GetProviders())
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
	resp, err := s.ListAuditLogs(ctx, pb.ListAuditLogsRequest_builder{
		StartTime: proto.String(now.Add(-1 * time.Hour).Format(time.RFC3339)),
		EndTime:   proto.String(now.Add(1 * time.Hour).Format(time.RFC3339)),
	}.Build())
	require.NoError(t, err)
	require.Len(t, resp.GetEntries(), 1)
	assert.Equal(t, "test-tool", resp.GetEntries()[0].GetToolName())
	assert.Equal(t, "user1", resp.GetEntries()[0].GetUserId())

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
	_, err = s.ListAuditLogs(ctx, pb.ListAuditLogsRequest_builder{StartTime: proto.String("invalid")}.Build())
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = s.ListAuditLogs(ctx, pb.ListAuditLogsRequest_builder{EndTime: proto.String("invalid")}.Build())
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
	assert.Len(t, resp.GetServices(), 1)
	assert.Equal(t, "svc_fallback", resp.GetServices()[0].GetName())
	assert.Equal(t, "OK", resp.GetServiceStates()[0].GetStatus())
}
