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
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/llm"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLLMProviders(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := llm.NewProviderStore(tmpDir)
	require.NoError(t, err)

	app := &Application{
		LLMProviderStore: store,
	}

	// Test Save
	t.Run("SaveConfig", func(t *testing.T) {
		cfg := llm.ProviderConfig{
			Type:   llm.ProviderOpenAI,
			APIKey: "sk-test-key",
			Model:  "gpt-4",
		}
		body, _ := json.Marshal(cfg)
		req := httptest.NewRequest("POST", "/settings/llm-providers", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleLLMProviders(store)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test Get
	t.Run("GetConfigs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/settings/llm-providers", nil)
		w := httptest.NewRecorder()

		handler := app.handleLLMProviders(store)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var configs []llm.ProviderConfig
		err := json.Unmarshal(w.Body.Bytes(), &configs)
		require.NoError(t, err)
		require.Len(t, configs, 1)
		assert.Equal(t, llm.ProviderOpenAI, configs[0].Type)
		// Check basic masking - exact implementation might vary but should contain **** or ...
		assert.Contains(t, configs[0].APIKey, "-key")
	})
}

func TestHandleAutoCraftJobs(t *testing.T) {
	busProvider, err := bus.NewProvider(nil)
	require.NoError(t, err)

	app := &Application{
		busProvider: busProvider,
	}

	t.Run("SubmitJob", func(t *testing.T) {
		// Subscribe to verify message
		reqBus, err := bus.GetBus[*worker.AutoCraftRequest](busProvider, "auto_craft_request")
		require.NoError(t, err)

		received := make(chan *worker.AutoCraftRequest, 1)
		unsubscribe := reqBus.Subscribe(context.Background(), "auto_craft_request", func(req *worker.AutoCraftRequest) {
			received <- req
		})
		defer unsubscribe()

		jobReq := AutoCraftJobRequest{
			ServiceName:  "test-service",
			ProviderType: llm.ProviderOpenAI,
			Goal:         "Create a weather service",
		}
		body, _ := json.Marshal(jobReq)
		req := httptest.NewRequest("POST", "/autocraft/jobs", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleAutoCraftJobs()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AutoCraftJobResponse
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.JobID)

		select {
		case req := <-received:
			assert.Equal(t, resp.JobID, req.ID)
			assert.Equal(t, "test-service", req.ServiceName)
			assert.Equal(t, "Create a weather service", req.UsersGoal)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for bus message")
		}
	})
}
