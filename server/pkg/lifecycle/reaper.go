package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LeaseStatus represents the public LeaseStatus entity.
//
// Summary: Defines the structured data model representing a status.
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
type LeaseStatus string

const (
	// StatusActive represents an active and valid lease.
	StatusActive LeaseStatus = "ACTIVE"
	// StatusExpired represents a lease that has passed its expiration time.
	StatusExpired LeaseStatus = "EXPIRED"
	// StatusPruned represents a lease that has been manually invalidated or cleaned up.
	StatusPruned LeaseStatus = "PRUNED"
)

// Lease represents the public Lease entity.
//
// Summary: Defines the structured data model representing a .
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
type Lease struct {
	IntentID           string
	SubagentSessionIDs []string
	Expiry             time.Time
	Status             LeaseStatus
	mu                 sync.RWMutex
}

// SubagentReaper represents the public SubagentReaper entity.
//
// Summary: Defines the structured data model representing a reaper.
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
type SubagentReaper struct {
	registry map[string]*Lease
	mu       sync.RWMutex
	ticker   *time.Ticker
	quit     chan struct{}
}

// NewSubagentReaper serves as a public interface for interacting with NewSubagentReaper.
//
// Summary: Constructs and returns an initialized subagent reaper ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewSubagentReaper() *SubagentReaper {
	return &SubagentReaper{
		registry: make(map[string]*Lease),
		quit:     make(chan struct{}),
	}
}

// RegisterBranch serves as a public interface for interacting with RegisterBranch.
//
// Summary: Register the branch appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *SubagentReaper) RegisterBranch(intentID string, ttl time.Duration) *Lease {
	r.mu.Lock()
	defer r.mu.Unlock()

	lease := &Lease{
		IntentID:           intentID,
		SubagentSessionIDs: []string{},
		Expiry:             time.Now().Add(ttl),
		Status:             StatusActive,
	}
	r.registry[intentID] = lease
	return lease
}

// RegisterSubagent serves as a public interface for interacting with RegisterSubagent.
//
// Summary: Register the subagent appropriately based on current system conditions.
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
func (r *SubagentReaper) RegisterSubagent(intentID string, sessionID string) error {
	r.mu.RLock()
	lease, exists := r.registry[intentID]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("lease not found for intent: %s", intentID)
	}

	lease.mu.Lock()
	defer lease.mu.Unlock()

	if lease.Status != StatusActive {
		return fmt.Errorf("cannot register subagent: lease is %s", lease.Status)
	}

	lease.SubagentSessionIDs = append(lease.SubagentSessionIDs, sessionID)
	return nil
}

// RecordHeartbeat serves as a public interface for interacting with RecordHeartbeat.
//
// Summary: Record the heartbeat appropriately based on current system conditions.
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
func (r *SubagentReaper) RecordHeartbeat(intentID string, signature string, extendBy time.Duration) error {
	r.mu.RLock()
	lease, exists := r.registry[intentID]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("lease not found for intent: %s", intentID)
	}

	lease.mu.Lock()
	defer lease.mu.Unlock()

	if lease.Status != StatusActive {
		return fmt.Errorf("cannot record heartbeat: lease is %s", lease.Status)
	}

	// In a real implementation, verify the signature here

	lease.Expiry = time.Now().Add(extendBy)
	return nil
}

// PruneIntent serves as a public interface for interacting with PruneIntent.
//
// Summary: Prune the intent appropriately based on current system conditions.
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
func (r *SubagentReaper) PruneIntent(intentID string) error {
	r.mu.RLock()
	lease, exists := r.registry[intentID]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("lease not found for intent: %s", intentID)
	}

	lease.mu.Lock()
	defer lease.mu.Unlock()

	lease.Status = StatusPruned
	// In a real implementation: terminate subagent connections and roll back state
	return nil
}

// Start serves as a public interface for interacting with Start.
//
// Summary: Start the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *SubagentReaper) Start(ctx context.Context, interval time.Duration) {
	r.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.sweep()
			case <-r.quit:
				r.ticker.Stop()
				return
			case <-ctx.Done():
				r.ticker.Stop()
				return
			}
		}
	}()
}

// Stop serves as a public interface for interacting with Stop.
//
// Summary: Stop the  appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *SubagentReaper) Stop() {
	close(r.quit)
}

// sweep checks for expired leases and marks them.
func (r *SubagentReaper) sweep() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, lease := range r.registry {
		lease.mu.Lock()
		if lease.Status == StatusActive && now.After(lease.Expiry) {
			lease.Status = StatusExpired
			// In a real implementation: terminate subagent connections and roll back state
		}
		lease.mu.Unlock()
	}
}

// GetLeaseStatus serves as a public interface for interacting with GetLeaseStatus.
//
// Summary: Fetches and returns the underlying lease status from the system state.
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
func (r *SubagentReaper) GetLeaseStatus(intentID string) (LeaseStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lease, exists := r.registry[intentID]
	if !exists {
		return "", fmt.Errorf("lease not found for intent: %s", intentID)
	}

	lease.mu.RLock()
	defer lease.mu.RUnlock()

	return lease.Status, nil
}
