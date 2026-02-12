package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	// Create temporary DB
	tmpDir, err := os.MkdirTemp("", "mcpany-sqlite-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := tmpDir + "/test.db"
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()

	t.Run("SaveMetric", func(t *testing.T) {
		metric := &storage.Metric{
			ServiceID: "service-1",
			Timestamp: time.Now().UnixMilli(),
			Status:    "UP",
		}
		err := store.SaveMetric(ctx, metric)
		assert.NoError(t, err)
	})

	t.Run("ListMetricHistory", func(t *testing.T) {
		// Add some metrics
		now := time.Now()
		m1 := &storage.Metric{ServiceID: "svc-a", Timestamp: now.Add(-1 * time.Hour).UnixMilli(), Status: "UP"}
		m2 := &storage.Metric{ServiceID: "svc-a", Timestamp: now.Add(-30 * time.Minute).UnixMilli(), Status: "DOWN"}
		m3 := &storage.Metric{ServiceID: "svc-b", Timestamp: now.Add(-10 * time.Minute).UnixMilli(), Status: "UP"}

		require.NoError(t, store.SaveMetric(ctx, m1))
		require.NoError(t, store.SaveMetric(ctx, m2))
		require.NoError(t, store.SaveMetric(ctx, m3))

		// Query
		history, err := store.ListMetricHistory(ctx, now.Add(-2*time.Hour))
		assert.NoError(t, err)
		assert.Len(t, history, 3) // svc-a (2 points) + svc-b (1 point) + previous test point (service-1)
		// Wait, service-1 was added in previous test step but same DB.
		// svc-a has 2
		assert.Len(t, history["svc-a"], 2)
		assert.Len(t, history["svc-b"], 1)

		assert.Equal(t, "UP", history["svc-a"][0].Status)
		assert.Equal(t, "DOWN", history["svc-a"][1].Status)
	})
}
