package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
)

func TestUserPreferencesPersistenceE2E(t *testing.T) {
	ctx := context.Background()

	// 1. Setup in-memory sqlite store
	store, err := sqlite.NewSqliteStorage(":memory:")
	require.NoError(t, err)
	defer store.Close()

	require.NoError(t, store.Init(ctx))

	application := &app.Application{
		Storage: store,
	}

	// 2. Mock a user in context
	userID := "test-admin"
	reqCtx := auth.ContextWithUser(ctx, userID)

	// 3. Test GET empty preferences
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
	req1 = req1.WithContext(reqCtx)
	rr1 := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock handler mapping
	})
	handler.ServeHTTP(rr1, req1)
}
