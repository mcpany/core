package integration

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/require"
)

func TestDashboardPersistence(t *testing.T) {
	// Start server (in-process)
	serverInfo := StartInProcessMCPANYServer(t, "DashboardTest", "test-key")
	defer serverInfo.CleanupFunc()

	// Create Dashboard Client
	client := pb.NewDashboardServiceClient(serverInfo.GRPCRegConn)

	ctx := context.Background()

	// 1. Save Layout
	layout := `[{"id":"widget-1"}]`
	_, err := client.SaveDashboardLayout(ctx, &pb.SaveDashboardLayoutRequest{
		LayoutJson: layout,
	})
	require.NoError(t, err)

	// 2. Get Layout
	resp, err := client.GetDashboardLayout(ctx, &pb.GetDashboardLayoutRequest{})
	require.NoError(t, err)
	require.Equal(t, layout, resp.GetLayoutJson())

	// 3. Update Layout
	layout2 := `[{"id":"widget-2"}]`
	_, err = client.SaveDashboardLayout(ctx, &pb.SaveDashboardLayoutRequest{
		LayoutJson: layout2,
	})
	require.NoError(t, err)

	resp2, err := client.GetDashboardLayout(ctx, &pb.GetDashboardLayoutRequest{})
	require.NoError(t, err)
	require.Equal(t, layout2, resp2.GetLayoutJson())
}
