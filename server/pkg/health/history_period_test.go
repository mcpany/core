package health

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPeriodicHistoryRecording(t *testing.T) {
	// Override heartbeat interval for testing
	originalInterval := heartbeatInterval
	heartbeatInterval = 100 * time.Millisecond

	// Disable cache to allow manual triggering of checks
	originalCacheDuration := cacheDuration
	cacheDuration = 0

	defer func() {
		heartbeatInterval = originalInterval
		cacheDuration = originalCacheDuration
	}()

	// Reset history store for this test
	historyMu.Lock()
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu.Unlock()

	svcName := "periodic-test-service"

	// Use builder to create config (Opaque API)
	cfg := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(svcName),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"), // Should succeed
            // No HealthCheck config implies manual check or default?
            // But getHealthCheckConfig returns interval=0 if HealthCheck is nil.
            // So NewChecker uses WithCheck (passive).
            // This might be easier to trigger manually?
		}.Build(),
	}.Build()

	checker := NewChecker(cfg)
	assert.NotNil(t, checker)

	// Start the checker (it might run in background too)
	checker.Start()

	// Manual trigger to force checks and verify heartbeat logic
	// We expect points roughly every 100ms.
	// We run for ~400ms total.

	for i := 0; i < 8; i++ {
		_ = checker.Check(context.Background())
		time.Sleep(50 * time.Millisecond)
	}

	checker.Stop()

	// Verify history
	history := GetHealthHistory()
	points, ok := history[svcName]
	assert.True(t, ok, "History should exist for service")

	// We expect roughly 4-5 points (initial + every 100ms)
	// If we have 400ms duration, and 100ms heartbeat.
	// Points at 0, 100, 200, 300, 400.

	assert.Greater(t, len(points), 1, "Should have more than 1 history point due to periodic heartbeat. Found: %d", len(points))

	if len(points) > 1 {
		// Verify intervals
		for i := 1; i < len(points); i++ {
			diff := points[i].Timestamp - points[i-1].Timestamp
			// Timestamp is millis.
			// Diff should be roughly 100ms (heartbeat interval).
			// Since we check every 50ms, the precision is +/- 50ms.
			// So expected diff >= 100ms - margin.
			// Actually, logic is: if time.Since(last) > heartbeat.
			// So it triggers on the *next* check after heartbeat interval.
			// So interval will be >= 100ms.
			assert.GreaterOrEqual(t, diff, int64(90), "Interval should be at least heartbeat interval (roughly)")
		}
	}
}
