// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleDebugSeed(t *testing.T) {
	// Setup application with memory storage
	store := memory.NewStore()
	app := &Application{
		Storage:     store,
		configPaths: []string{}, // Empty paths to avoid FS usage
	}

	// Pre-populate some data to verify it gets cleared
	ctx := context.Background()
	existingService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("existing-service"),
	}.Build()
	require.NoError(t, store.SaveService(ctx, existingService))

	// Define seed payload
	newService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("new-service"),
	}.Build()

	svcBytes, err := protojson.Marshal(newService)
	require.NoError(t, err)

	seedReq := SeedRequest{
		ServicesRaw: []json.RawMessage{svcBytes},
	}
	body, err := json.Marshal(seedReq)
	require.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debug/seed", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Invoke handler
	app.handleDebugSeed().ServeHTTP(w, req)

	// Verify response
	require.Equal(t, http.StatusOK, w.Code)

	// Verify existing data is gone
	svc, err := store.GetService(ctx, "existing-service")
	require.NoError(t, err)
	require.Nil(t, svc)

	// Verify new data is present
	svc, err = store.GetService(ctx, "new-service")
	require.NoError(t, err)
	require.NotNil(t, svc)
	require.Equal(t, "new-service", svc.GetName())
}
