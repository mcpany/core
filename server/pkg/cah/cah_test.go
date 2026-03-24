package cah

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockMonitor struct {
	id         string
	shouldFail bool
	delay      time.Duration
}

func (m *mockMonitor) ValidateRequest(ctx context.Context, requestID string, intent string, payload []byte) (string, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.shouldFail {
		return "", errors.New("monitor rejected")
	}
	return "sig_" + m.id, nil
}

func (m *mockMonitor) ID() string {
	return m.id
}

func TestCAHAdapter_New(t *testing.T) {
	monitors := []MonitorAgent{
		&mockMonitor{id: "m1"},
		&mockMonitor{id: "m2"},
	}

	adapter, err := NewCAHAdapter(monitors, 2, time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, adapter)

	// Test invalid threshold (too low)
	_, err = NewCAHAdapter(monitors, 0, time.Second)
	assert.Error(t, err)

	// Test invalid threshold (too high)
	_, err = NewCAHAdapter(monitors, 3, time.Second)
	assert.Error(t, err)
}

func TestCAHAdapter_ValidateWithQuorum(t *testing.T) {
	monitors := []MonitorAgent{
		&mockMonitor{id: "m1"},
		&mockMonitor{id: "m2"},
		&mockMonitor{id: "m3"},
	}

	adapter, err := NewCAHAdapter(monitors, 2, 2*time.Second)
	assert.NoError(t, err)

	ctx := context.Background()
	requestID := "req_123"
	intent := "read_file"
	payload := []byte("payload_data")

	// Test successful quorum
	sigs, err := adapter.ValidateWithQuorum(ctx, requestID, intent, payload)
	assert.NoError(t, err)
	assert.Len(t, sigs, 2) // quorum size is 2, it returns immediately

	// Test rejection by some monitors (quorum not met)
	monitors2 := []MonitorAgent{
		&mockMonitor{id: "m1"},
		&mockMonitor{id: "m2", shouldFail: true},
		&mockMonitor{id: "m3", shouldFail: true},
	}
	adapter2, _ := NewCAHAdapter(monitors2, 2, 2*time.Second)

	sigs, err = adapter2.ValidateWithQuorum(ctx, requestID, intent, payload)
	assert.Error(t, err)
	assert.Nil(t, sigs)
	assert.Contains(t, err.Error(), "cah quorum rejected request")

	// Test timeout
	monitors3 := []MonitorAgent{
		&mockMonitor{id: "m1", delay: 3 * time.Second},
	}
	adapter3, _ := NewCAHAdapter(monitors3, 1, 100*time.Millisecond) // Short timeout

	sigs, err = adapter3.ValidateWithQuorum(ctx, requestID, intent, payload)
	assert.Error(t, err)
	assert.Nil(t, sigs)
	assert.Contains(t, err.Error(), "validation timeout exceeded")
}
