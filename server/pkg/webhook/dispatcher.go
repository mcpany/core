// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/storage"
)

// Dispatcher handles the dispatching of system events to registered webhooks.
type Dispatcher struct {
	storage storage.Storage
	bp      *bus.Provider
	client  *http.Client
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(s storage.Storage, bp *bus.Provider) *Dispatcher {
	return &Dispatcher{
		storage: s,
		bp:      bp,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Start subscribes to relevant bus topics and starts listening for events.
func (d *Dispatcher) Start(ctx context.Context) error {
	// Subscribe to ServiceRegistrationResult
	srrBus, err := bus.GetBus[*bus.ServiceRegistrationResult](d.bp, bus.ServiceRegistrationResultTopic)
	if err != nil {
		return err
	}
	srrBus.Subscribe(ctx, bus.ServiceRegistrationResultTopic, func(msg *bus.ServiceRegistrationResult) {
		d.dispatch(ctx, "service.registered", msg)
	})

	// Subscribe to ToolExecutionResult
	terBus, err := bus.GetBus[*bus.ToolExecutionResult](d.bp, bus.ToolExecutionResultTopic)
	if err != nil {
		return err
	}
	terBus.Subscribe(ctx, bus.ToolExecutionResultTopic, func(msg *bus.ToolExecutionResult) {
		d.dispatch(ctx, "tool.invoked", msg)
	})

	return nil
}

func (d *Dispatcher) dispatch(ctx context.Context, eventType string, payload any) {
	// Use a detached context for the database list operation to prevent cancellation
	// from interrupting the dispatch flow if the parent context (e.g. request) is cancelled.
	// However, usually Start() context is long-lived (server lifecycle).
	webhooks, err := d.storage.ListSystemWebhooks(ctx)
	if err != nil {
		// Log error (should use a logger)
		fmt.Printf("Failed to list webhooks: %v\n", err)
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	for _, wh := range webhooks {
		if !wh.GetActive() {
			continue
		}

		match := false
		for _, e := range wh.GetEvents() {
			if e == "all" || e == eventType {
				match = true
				break
			}
		}
		if !match {
			continue
		}

		// Fire and forget
		go d.send(wh, eventType, payloadBytes)
	}
}

func (d *Dispatcher) send(wh *configv1.SystemWebhook, eventType string, payload []byte) {
	req, err := http.NewRequest("POST", wh.GetUrl(), bytes.NewBuffer(payload))
	if err != nil {
		d.updateStatus(wh, "failure", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MCP-Event", eventType)
	req.Header.Set("X-MCP-Webhook-ID", wh.GetId())

	if wh.GetSecret() != "" {
		mac := hmac.New(sha256.New, []byte(wh.GetSecret()))
		mac.Write(payload)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-MCP-Signature", "sha256="+signature)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		d.updateStatus(wh, "failure", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		d.updateStatus(wh, "success", "")
	} else {
		d.updateStatus(wh, "failure", fmt.Sprintf("Status %d", resp.StatusCode))
	}
}

func (d *Dispatcher) updateStatus(wh *configv1.SystemWebhook, status, errorMsg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update fields
	wh.SetLastTriggeredAt(time.Now().Format(time.RFC3339))
	wh.SetLastStatus(status)
	wh.SetLastError(errorMsg)

	_ = d.storage.UpdateSystemWebhook(ctx, wh)
}
