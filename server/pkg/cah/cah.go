package cah

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// MonitorAgent represents a security/policy validator in the quorum.
//
// Summary: Evaluates a request and returns a cryptographically bound signature.
type MonitorAgent interface {
	// ValidateRequest evaluates a request and returns a cryptographically bound
	// signature if approved, or an error if rejected.
	ValidateRequest(ctx context.Context, requestID string, intent string, payload []byte) (string, error)
	// ID returns the unique identifier of the monitor agent.
	ID() string
}

// CAHAdapter acts as the central arbiter for verifying agent interactions.
//
// Summary: Manages a decentralized quorum of MonitorAgents to collect approvals.
type CAHAdapter struct {
	monitors        []MonitorAgent
	quorumThreshold int
	timeout         time.Duration
}

// NewCAHAdapter creates a new Cognitive Attestation Hub (CAH) Adapter.
//
// Summary: Initializes and returns a new CAHAdapter instance.
//
// Parameters:
//   - monitors: A list of MonitorAgent instances that form the quorum.
//   - threshold: The minimum number of approvals required to proceed.
//   - timeout: The maximum time to wait for quorum consensus.
//
// Returns:
//   - *CAHAdapter: A new CAHAdapter instance.
//   - error: An error if the threshold is greater than the number of monitors or less than 1.
//
// Errors:
//   - Returns an error if threshold is invalid.
//
// Side Effects:
//   - None.
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

// ValidateWithQuorum initiates a consensus gathering process for a given request.
//
// Summary: Starts a consensus gathering process among the monitor quorum.
//
// Parameters:
//   - ctx: The context for the request.
//   - requestID: A unique identifier for the request being validated.
//   - intent: The declared intent of the request (e.g., "read_file", "execute_command").
//   - payload: The serialized payload of the request.
//
// Returns:
//   - []string: A list of cryptographic signatures from the approving monitors.
//   - error: An error if quorum is not reached within the timeout.
//
// Errors:
//   - Returns an error if the context deadline is exceeded.
//   - Returns an error if the required number of approvals is not met.
//
// Side Effects:
//   - Interacts with all configured MonitorAgent instances.
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
