package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHistory(t *testing.T) {
	// Reset store
	historyMu.Lock()
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu.Unlock()

	AddHealthStatus("svc1", "healthy")
	AddHealthStatus("svc1", "unhealthy")

	hist := GetHealthHistory()
	assert.Len(t, hist["svc1"], 2)
	assert.Equal(t, "healthy", hist["svc1"][0].Status)
	assert.Equal(t, "unhealthy", hist["svc1"][1].Status)

	// Test pruning
	for i := 0; i < 1100; i++ {
		AddHealthStatus("svc2", "ok")
	}
	hist = GetHealthHistory()
	assert.Len(t, hist["svc2"], 1000)
}
