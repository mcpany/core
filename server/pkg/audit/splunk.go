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

const (
	splunkBufferSize = 1000
	splunkWorkers    = 2
	splunkBatchSize  = 10
	splunkBatchWait  = 1 * time.Second
)

// SplunkAuditStore sends audit logs to Splunk HTTP Event Collector. Summary: Asynchronous audit store that pushes logs to Splunk via HEC.
//
// Summary: SplunkAuditStore sends audit logs to Splunk HTTP Event Collector. Summary: Asynchronous audit store that pushes logs to Splunk via HEC.
//
// Fields:
//   - Contains the configuration and state properties required for SplunkAuditStore functionality.
type SplunkAuditStore struct {
	config *configv1.SplunkConfig
	client *http.Client
	queue  chan Entry
	wg     sync.WaitGroup
	done   chan struct{}
}

// NewSplunkAuditStore creates a new SplunkAuditStore.
//
// Summary: Initializes a new SplunkAuditStore with background workers.
//
// Parameters:
//   - config: *configv1.SplunkConfig. The Splunk HEC configuration.
//
// Returns:
//   - *SplunkAuditStore: The initialized store.
//
// Side Effects:
//   - Starts background workers.
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

	resp, err := e.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send batch to splunk: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Splunk HEC returned status: %d\n", resp.StatusCode)
	}
}


// Read implements the Store interface. Summary: Reads audit entries (Not implemented). Parameters: - _: context.Context. Unused. - _: Filter. Unused. Returns: - []Entry: Nil. - error: Always returns "not implemented".
//
// Summary: Read implements the Store interface. Summary: Reads audit entries (Not implemented). Parameters: - _: context.Context. Unused. - _: Filter. Unused. Returns: - []Entry: Nil. - error: Always returns "not implemented".
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - _ (Filter): The _ parameter used in the operation.
//
// Returns:
//   - ([]Entry): The resulting []Entry object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (e *SplunkAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for splunk audit store")
}

// Close closes the queue and waits for workers to finish. Summary: Shuts down the Splunk audit store. Returns: - error: Always nil. Side Effects: - Closes channels. - Flushes pending batches.
//
// Summary: Close closes the queue and waits for workers to finish. Summary: Shuts down the Splunk audit store. Returns: - error: Always nil. Side Effects: - Closes channels. - Flushes pending batches.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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
