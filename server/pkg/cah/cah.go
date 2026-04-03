package cah

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// MonitorAgent represents the public MonitorAgent entity.
//
// Summary: Defines the structured data model representing a agent.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MonitorAgent interface {
	// ValidateRequest evaluates a request and returns a cryptographically bound
	// signature if approved, or an error if rejected.
	ValidateRequest(ctx context.Context, requestID string, intent string, payload []byte) (string, error)
	// ID returns the unique identifier of the monitor agent.
	ID() string
}

// CAHAdapter represents the public CAHAdapter entity.
//
// Summary: Defines the structured data model representing a adapter.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type CAHAdapter struct {
	monitors        []MonitorAgent
	quorumThreshold int
	timeout         time.Duration
}

// NewCAHAdapter serves as a public interface for interacting with NewCAHAdapter.
//
// Summary: Constructs and returns an initialized cah adapter ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewCAHAdapter(monitors []MonitorAgent, threshold int, timeout time.Duration) (*CAHAdapter, error) {
	if threshold < 1 {
		return nil, fmt.Errorf("quorum threshold must be at least 1")
	}
	if threshold > len(monitors) {
		return nil, fmt.Errorf("quorum threshold cannot exceed the number of monitors (%d > %d)", threshold, len(monitors))
	}

	return &CAHAdapter{
		monitors:        monitors,
		quorumThreshold: threshold,
		timeout:         timeout,
	}, nil
}

// ValidateWithQuorum serves as a public interface for interacting with ValidateWithQuorum.
//
// Summary: Validate the with quorum appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *CAHAdapter) ValidateWithQuorum(ctx context.Context, requestID string, intent string, payload []byte) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var signatures []string
	var rejections []error

	// Channels to signal completion or early return
	sigChan := make(chan string, len(c.monitors))
	errChan := make(chan error, len(c.monitors))

	// Initiate validation with all monitors concurrently
	for _, monitor := range c.monitors {
		wg.Add(1)
		go func(m MonitorAgent) {
			defer wg.Done()

			sig, err := m.ValidateRequest(ctx, requestID, intent, payload)
			if err != nil {
				logging.GetLogger().WarnContext(ctx, "Monitor agent rejected request", "monitor_id", m.ID(), "request_id", requestID, "error", err)
				errChan <- err
			} else {
				sigChan <- sig
			}
		}(monitor)
	}

	// Wait for all monitors to respond
	go func() {
		wg.Wait()
		close(sigChan)
		close(errChan)
	}()

	// Loop to process results
	for {
		select {
		case sig, ok := <-sigChan:
			if ok {
				mu.Lock()
				signatures = append(signatures, sig)
				if len(signatures) >= c.quorumThreshold {
					mu.Unlock()
					logging.GetLogger().InfoContext(ctx, "CAH quorum reached", "request_id", requestID, "approvals", len(signatures), "threshold", c.quorumThreshold)
					return signatures, nil
				}
				mu.Unlock()
			}
		case err, ok := <-errChan:
			if ok {
				mu.Lock()
				rejections = append(rejections, err)
				if len(rejections) > len(c.monitors)-c.quorumThreshold {
					mu.Unlock()
					return nil, fmt.Errorf("cah quorum rejected request: got %d approvals, needed %d (rejections: %v)", len(signatures), c.quorumThreshold, rejections)
				}
				mu.Unlock()
			}
		case <-ctx.Done():
			mu.Lock()
			defer mu.Unlock()
			return nil, fmt.Errorf("cah validation timeout exceeded waiting for quorum (got %d/%d approvals)", len(signatures), c.quorumThreshold)
		}
	}
}
