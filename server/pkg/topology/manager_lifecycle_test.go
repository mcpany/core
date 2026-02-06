// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestManager_SessionCleanup verifies that inactive sessions are removed.
func TestManager_SessionCleanup(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	// Mock time
	currTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m.nowFunc = func() time.Time {
		return currTime
	}

	// 1. Create a session
	m.RecordActivity("session-active", nil, 10*time.Millisecond, false, "")
	m.RecordActivity("session-stale", nil, 10*time.Millisecond, false, "")

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, ok1 := m.sessions["session-active"]
		_, ok2 := m.sessions["session-stale"]
		m.mu.RUnlock()
		return ok1 && ok2
	}, 1*time.Second, 10*time.Millisecond)

	// 2. Advance time by 25 hours
	currTime = currTime.Add(25 * time.Hour)

	// 3. Keep "session-active" alive
	m.RecordActivity("session-active", nil, 10*time.Millisecond, false, "")

	// 4. Trigger cleanup (needs 100 requests on a session)
	// We'll spam requests on "session-active" to trigger the check.
	// Since the check logic is: if session.RequestCount%100 == 0
	// We need to hit that condition.
	for i := 0; i < 110; i++ {
		m.RecordActivity("session-active", nil, 1*time.Millisecond, false, "")
		// Small sleep to allow channel processing
		if i%10 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	// 5. Verify "session-stale" is gone
	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["session-stale"]
		m.mu.RUnlock()
		return !exists // Should be false (deleted)
	}, 2*time.Second, 100*time.Millisecond, "Session should be cleaned up")

	// Verify active session still exists
	m.mu.RLock()
	_, exists := m.sessions["session-active"]
	m.mu.RUnlock()
	assert.True(t, exists, "Active session should remain")
}

// TestManager_TrafficHistoryCleanup verifies that old traffic history is removed.
func TestManager_TrafficHistoryCleanup(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	// Mock time
	startTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	currTime := startTime
	m.nowFunc = func() time.Time {
		return currTime
	}

	// 1. Create old history (25 hours ago)
	// We can't directly inject into trafficHistory map easily as it uses unix timestamp keys based on nowFunc.
	// Instead we rely on RecordActivity using nowFunc.

	// Set time to 25h ago
	oldTime := startTime.Add(-25 * time.Hour)
	currTime = oldTime
	m.RecordActivity("session-1", nil, 10*time.Millisecond, false, "")

	// Wait for it to be recorded
	assert.Eventually(t, func() bool {
		m.mu.RLock()
		defer m.mu.RUnlock()
		key := oldTime.Truncate(time.Minute).Unix()
		_, ok := m.trafficHistory[key]
		return ok
	}, 1*time.Second, 10*time.Millisecond)

	// 2. Set time to now
	currTime = startTime

	// 3. Trigger cleanup
	for i := 0; i < 110; i++ {
		m.RecordActivity("session-1", nil, 1*time.Millisecond, false, "")
		if i%10 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	// 4. Verify old history is gone
	assert.Eventually(t, func() bool {
		m.mu.RLock()
		defer m.mu.RUnlock()
		key := oldTime.Truncate(time.Minute).Unix()
		_, ok := m.trafficHistory[key]
		return !ok
	}, 2*time.Second, 100*time.Millisecond, "Old traffic history should be cleaned up")
}

func TestManager_Concurrency_Safe(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	// Basic mocks setup needed for GetGraph
	mockTM.On("ListTools").Return([]tool.Tool{})
	svcConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svcConfig}, nil)

	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	var wg sync.WaitGroup
	startCh := make(chan struct{})

	// Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-startCh
		for i := 0; i < 1000; i++ {
			sessionID := fmt.Sprintf("session-%d", i%10)
			m.RecordActivity(sessionID, nil, time.Duration(i)*time.Millisecond, false, "")
			time.Sleep(10 * time.Microsecond)
		}
	}()

	// Reader: GetStats
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-startCh
		for i := 0; i < 100; i++ {
			m.GetStats("")
			time.Sleep(100 * time.Microsecond)
		}
	}()

	// Reader: GetGraph
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-startCh
		for i := 0; i < 100; i++ {
			m.GetGraph(context.Background())
			time.Sleep(100 * time.Microsecond)
		}
	}()

	// Reader: GetTrafficHistory
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-startCh
		for i := 0; i < 100; i++ {
			m.GetTrafficHistory("")
			time.Sleep(100 * time.Microsecond)
		}
	}()

	close(startCh)
	wg.Wait()
}
