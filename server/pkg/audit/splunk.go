// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// NewSplunkAuditStore creates a new SplunkAuditStore.
//
// Summary: Initializes a new SplunkAuditStore with background workers.
//
// Parameters:
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - config: *configv1.SplunkConfig. The Splunk HEC configuration.
//
// Returns:
//   - *SplunkAuditStore: The initialized store.
//
// Side Effects:
//   - Starts background workers.
// Errors:
//   - triggers relevant error states on failure.
func NewSplunkAuditStore(config *configv1.SplunkConfig) *SplunkAuditStore {
	if config == nil {
		config = &configv1.SplunkConfig{}
	}
	store := &SplunkAuditStore{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		queue: make(chan Entry, splunkBufferSize),
		done:  make(chan struct{}),
	}

	for i := 0; i < splunkWorkers; i++ {
		store.wg.Add(1)
		go store.worker()
	}

	return store
}

func (e *SplunkAuditStore) worker() {
	defer e.wg.Done()
	var batch []Entry
	ticker := time.NewTicker(splunkBatchWait)
	defer ticker.Stop()

	for {
		select {
		case entry, ok := <-e.queue:
			if !ok {
				e.sendBatch(batch)
				return
			}
			batch = append(batch, entry)
			if len(batch) >= splunkBatchSize {
				e.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				e.sendBatch(batch)
				batch = nil
			}
		case <-e.done:
			// Drain queue
			for entry := range e.queue {
				batch = append(batch, entry)
				if len(batch) >= splunkBatchSize {
					e.sendBatch(batch)
					batch = nil
				}
			}
			e.sendBatch(batch)
			return
		}
	}
}

// Write implements the Store interface.
//
// Summary: Queues an audit entry for sending to Splunk.
//
// Parameters:
//   - _: context.Context. Unused.
//   - entry: Entry. The audit entry.
//
// Returns:
//   - error: An error if the queue is full.
//
// Errors:
//   - Returns "audit queue full" if the buffer is exhausted.
//
// Side Effects:
//   - Sends entry to a buffered channel.
func (e *SplunkAuditStore) Write(_ context.Context, entry Entry) error {
	select {
	case e.queue <- entry:
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Splunk audit queue full, dropping log: %s\n", entry.ToolName)
		return fmt.Errorf("audit queue full")
	}
}

func (e *SplunkAuditStore) sendBatch(batch []Entry) {
	if len(batch) == 0 {
		return
	}

	buf := new(bytes.Buffer)
	for _, entry := range batch {
		event := map[string]interface{}{
			"time":       entry.Timestamp.Unix(),
			"host":       "mcpany",
			"source":     e.config.GetSource(),
			"sourcetype": e.config.GetSourcetype(),
			"index":      e.config.GetIndex(),
			"event":      entry,
		}
		payload, err := json.Marshal(event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal splunk event: %v\n", err)
			continue
		}
		buf.Write(payload)
		buf.WriteString("\n")
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", e.config.GetHecUrl(), buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create splunk request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Splunk "+e.config.GetToken())
	req.Header.Set("Content-Type", "application/json")

// Read implements the Store interface.
//
// Summary: Reads audit entries (Not implemented).
//
// Close closes the queue and waits for workers to finish.
//
// Summary: Shuts down the Splunk audit store.
//
// Returns:
//   - error: Always nil.
//
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Side Effects:
//   - Closes channels.
//   - Flushes pending batches.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
func (e *SplunkAuditStore) Close() error {
	if e.done != nil {
		close(e.done)
	}
	if e.queue != nil {
		close(e.queue)
	}
	e.wg.Wait()
	return nil
}
