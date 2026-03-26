package cah

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// MonitorAgent represents a security/policy validator within the quorum system.
//
// Summary: Defines the contract for agents that validate and approve requests.
type MonitorAgent interface {
	// ValidateRequest evaluates a request and returns a cryptographically bound
	// signature if approved, or an error if rejected.
	ValidateRequest(ctx context.Context, requestID string, intent string, payload []byte) (string, error)
	// ID returns the unique identifier of the monitor agent.
	ID() string
}

// CAHAdapter acts as the central arbiter for verifying agent interactions.
// It manages a decentralized quorum of MonitorAgents to collect approvals.
//
// Summary: Manages the quorum validation process for the consensus agent.
type CAHAdapter struct {
	monitors        []MonitorAgent
	quorumThreshold int
	timeout         time.Duration
}

// NewCAHAdapter creates a new Cognitive Attestation Hub (CAH) Adapter.
//
// Summary: Initializes a new CAHAdapter with the given quorum threshold and monitors.
//
// Parameters:
//   - monitors: The set of monitor agents participating in the quorum.
//   - threshold: The required number of approvals to reach consensus.
//   - timeout: The maximum duration to wait for the quorum to respond.
//
// Returns:
//   - *CAHAdapter: The configured CAHAdapter instance.
//   - error: Returns an error if the threshold is invalid.
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
// Summary: Concurrently polls all monitor agents and waits for the required threshold of approvals.
//
// Parameters:
//   - ctx: The context governing the validation timeout.
//   - requestID: The unique identifier for the request being validated.
//   - intent: The declared action intent of the request.
//   - payload: The raw data payload of the request.
//
// Returns:
//   - []string: A list of cryptographic signatures from approving monitors.
//   - error: Returns an error if the quorum is not reached or if the context times out.
//
// Errors:
//   - Returns an error if the timeout is exceeded before reaching the threshold.
//   - Returns an error if too many monitors reject the request.
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
