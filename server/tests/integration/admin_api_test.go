package integration

import (
	"context"
	"net"
	"testing"

	"github.com/mcpany/core/pkg/admin"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

// TestAdminAPI_CreateDeleteService tests the end-to-end flow of adding and removing a service via the Admin API.
// Note: This test mocks the underlying factory/upstream to avoid external dependencies,
// but tests the full wiring of the Admin API -> Service Registry -> Manager.
func TestAdminAPI_CreateDeleteService(t *testing.T) {
	// 1. Setup Server Components
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)

	mockFactory := &MockFactory{}
	registry := serviceregistry.New(mockFactory, toolManager, nil, nil, nil)

	// Create Admin Server
	cache := middleware.NewCachingMiddleware(toolManager)
	adminServer := admin.NewServer(cache, toolManager, registry)

	// Start gRPC Server
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s := gogrpc.NewServer()
	pb_admin.RegisterAdminServiceServer(s, adminServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited: %v", err)
		}
	}()
	defer s.Stop()

	// 2. Setup Client
	conn, err := gogrpc.Dial(lis.Addr().String(), gogrpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb_admin.NewAdminServiceClient(conn)

	// 3. Test CreateService
	svcConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("e2e-test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://example.com"),
			},
		},
	}

	createResp, err := client.CreateService(ctx, &pb_admin.CreateServiceRequest{Service: svcConfig})
	require.NoError(t, err)
	assert.Equal(t, "e2e-test-service", createResp.GetServiceId())

	// Verify service is listed
	listResp, err := client.ListServices(ctx, &pb_admin.ListServicesRequest{})
	require.NoError(t, err)
	found := false
	for _, s := range listResp.GetServices() {
		if s.GetName() == "e2e-test-service" {
			found = true
			break
		}
	}
	assert.True(t, found, "Service should be listed after creation")

	// 4. Test DeleteService
	serviceID := "e2e-test-service"
	deleteResp, err := client.DeleteService(ctx, &pb_admin.DeleteServiceRequest{ServiceId: &serviceID})
	require.NoError(t, err)
	assert.NotNil(t, deleteResp)

	// Verify service is gone
	listResp, err = client.ListServices(ctx, &pb_admin.ListServicesRequest{})
	require.NoError(t, err)
	found = false
	for _, s := range listResp.GetServices() {
		if s.GetName() == "e2e-test-service" {
			found = true
			break
		}
	}
	assert.False(t, found, "Service should not be listed after deletion")
}

// MockFactory implements factory.Factory
type MockFactory struct{}

func (f *MockFactory) NewUpstream(cfg *configv1.UpstreamServiceConfig) (upstream.Upstream, error) {
	// Return a mock Upstream
	return &MockUpstream{}, nil
}

type MockUpstream struct{}

func (u *MockUpstream) Register(ctx context.Context, cfg *configv1.UpstreamServiceConfig, tm tool.ManagerInterface, pm prompt.ManagerInterface, rm resource.ManagerInterface, reRegister bool) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return cfg.GetName(), []*configv1.ToolDefinition{}, []*configv1.ResourceDefinition{}, nil
}

func (u *MockUpstream) Shutdown(ctx context.Context) error {
	return nil
}
