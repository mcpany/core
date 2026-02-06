// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/llm"
)

func TestAutoCraftWorker(t *testing.T) {
	// Setup Bus (defaults to InMemory)
	busProvider, err := bus.NewProvider(nil)
	if err != nil {
		t.Fatalf("Failed to create bus provider: %v", err)
	}

	// Setup LLM Store
	tempDir, err := os.MkdirTemp("", "llm_store_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := llm.NewProviderStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Add a dummy key for OpenAI
	err = store.SaveConfig(llm.ProviderConfig{
		Type:   llm.ProviderOpenAI,
		APIKey: "sk-test-key",
	})
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Initialize Worker
	w := New(busProvider, &Config{MaxWorkers: 1, MaxQueueSize: 10})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start AutoCraftWorker
	w.StartAutoCraftWorker(ctx, store)

	// Subscribe to results
	resBus, err := bus.GetBus[*AutoCraftResult](busProvider, "auto_craft_result")
	if err != nil {
		t.Fatalf("Failed to get result bus: %v", err)
	}

	resultChan := make(chan *AutoCraftResult, 10)
	resBus.Subscribe(ctx, "auto_craft_result", func(res *AutoCraftResult) {
		resultChan <- res
	})

	// Publish Request
	reqBus, err := bus.GetBus[*AutoCraftRequest](busProvider, "auto_craft_request")
	if err != nil {
		t.Fatalf("Failed to get request bus: %v", err)
	}

	req := &AutoCraftRequest{
		ID:           "test-job-1",
		UsersGoal:    "Create a weather service",
		ServiceName:  "weather-service",
		ProviderType: llm.ProviderOpenAI,
	}
	// manually set CID since BaseMessage might not handle it automatically in struct literal if embedded
	req.SetCorrelationID("test-job-1")

	// Wait for worker to subscribe
	time.Sleep(100 * time.Millisecond)

	err = reqBus.Publish(ctx, "auto_craft_request", req)
	if err != nil {
		t.Fatalf("Failed to publish request: %v", err)
	}

	// Verify Results
	// Expect "in_progress" then "success"
	timeout := time.After(5 * time.Second)
	var finalResult *AutoCraftResult

	for {
		select {
		case res := <-resultChan:
			if res.CID != "test-job-1" {
				continue
			}
			t.Logf("Received status: %s", res.Status)
			if res.Status == "success" {
				finalResult = res
				goto Done
			}
			if res.Status == "failed" {
				t.Fatalf("Job failed unexpectedly: %s", res.ErrorMessage)
			}
		case <-timeout:
			t.Fatal("Timeout waiting for job completion")
		}
	}

Done:
	if finalResult == nil {
		t.Fatal("Did not receive success result")
	}
	if finalResult.ConfigJSON == "" {
		t.Error("ConfigJSON shouldn't be empty")
	}
}

func TestAutoCraftWorker_NoKey(t *testing.T) {
	// Setup Bus (defaults to InMemory)
	busProvider, err := bus.NewProvider(nil)
	if err != nil {
		t.Fatalf("Failed to create bus provider: %v", err)
	}

	// Setup LLM Store (Empty)
	tempDir, err := os.MkdirTemp("", "llm_store_test_nokey")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := llm.NewProviderStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Initialize Worker
	w := New(busProvider, &Config{MaxWorkers: 1, MaxQueueSize: 10})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start AutoCraftWorker
	w.StartAutoCraftWorker(ctx, store)

	// Subscribe to results
	resBus, err := bus.GetBus[*AutoCraftResult](busProvider, "auto_craft_result")
	if err != nil {
		t.Fatalf("Failed to get result bus: %v", err)
	}

	resultChan := make(chan *AutoCraftResult, 10)
	resBus.Subscribe(ctx, "auto_craft_result", func(res *AutoCraftResult) {
		resultChan <- res
	})

	// Publish Request for Missing Provider
	reqBus, err := bus.GetBus[*AutoCraftRequest](busProvider, "auto_craft_request")
	if err != nil {
		t.Fatalf("Failed to get request bus: %v", err)
	}

	req := &AutoCraftRequest{
		ID:           "test-job-2",
		UsersGoal:    "Create a weather service",
		ProviderType: llm.ProviderGemini, // We didn't add this key
	}
	req.SetCorrelationID("test-job-2")

	// Wait for worker to subscribe
	time.Sleep(100 * time.Millisecond)

	err = reqBus.Publish(ctx, "auto_craft_request", req)
	if err != nil {
		t.Fatalf("Failed to publish request: %v", err)
	}

	// Verify Results
	timeout := time.After(5 * time.Second)
	select {
	case res := <-resultChan:
		if res.CID == "test-job-2" && res.Status == "failed" {
			t.Log("Received expected failure")
			return
		}
	case <-timeout:
		t.Fatal("Timeout waiting for failure result")
	}
}
