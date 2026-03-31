package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LeaseStatus represents the current state of a lease.
type LeaseStatus string

const (
	// Summary: Defines the Active state of a lease.
	StatusActive  LeaseStatus = "ACTIVE"
	// Summary: Defines the Expired state of a lease.
	StatusExpired LeaseStatus = "EXPIRED"
	// Summary: Defines the Pruned state of a lease.
	StatusPruned  LeaseStatus = "PRUNED"
)

// Lease represents an intent-bound lease for subagents.
type Lease struct {
	IntentID           string
	SubagentSessionIDs []string
	Expiry             time.Time
	Status             LeaseStatus
	mu                 sync.RWMutex
}

// SubagentReaper manages the lifecycle of intent-bound subagents.
type SubagentReaper struct {
	registry map[string]*Lease
	mu       sync.RWMutex
	ticker   *time.Ticker
	quit     chan struct{}
}

// NewSubagentReaper initializes a new Active Subagent Reaper.
//
// Summary: Creates a new SubagentReaper instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *SubagentReaper: The initialized reaper.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Initializes internal channels and maps.
func NewSubagentReaper() *SubagentReaper {
	return &SubagentReaper{
		registry: make(map[string]*Lease),
		quit:     make(chan struct{}),
	}
}

// RegisterBranch creates a new speculative intent branch and assigns a lease.
//
// Summary: Registers a new intent branch with a lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//   - ttl (time.Duration): The time-to-live for the lease.
//
// Returns:
//   - *Lease: The newly created lease.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the internal registry map.
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

// RegisterSubagent attaches a subagent session to a lease.
//
// Summary: Registers a subagent with an existing intent lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//   - sessionID (string): The unique identifier for the subagent session.
//
// Returns:
//   - error: An error if the lease is not found or not active.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//   - Returns "cannot register subagent: lease is [status]" if not active.
//
// Side Effects:
//   - Modifies the lease's SubagentSessionIDs list.
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

// RecordHeartbeat updates the Last-Seen timestamp (Expiry) for an active lease.
//
// Summary: Extends the expiration time of an active lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//   - signature (string): The heartbeat signature.
//   - extendBy (time.Duration): The duration to extend the lease by.
//
// Returns:
//   - error: An error if the lease is not found or not active.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//   - Returns "cannot record heartbeat: lease is [status]" if not active.
//
// Side Effects:
//   - Modifies the Expiry timestamp of the lease.
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

// PruneIntent manually invalidates a lease and rolls back uncommitted writes.
//
// Summary: Marks an intent lease as pruned.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//
// Returns:
//   - error: An error if the lease is not found.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//
// Side Effects:
//   - Changes the lease status to StatusPruned.
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

// Start begins the Reaper Daemon background worker.
//
// Summary: Starts the background process to sweep expired leases.
//
// Parameters:
//   - ctx (context.Context): The context to control the daemon lifecycle.
//   - interval (time.Duration): The interval between sweep operations.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Starts a new goroutine for the daemon worker.
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

// Stop halts the Reaper Daemon.
//
// Summary: Stops the background process sweeping expired leases.
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
//   - Closes the quit channel to signal the daemon to stop.
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

// GetLeaseStatus returns the status of a given lease.
//
// Summary: Retrieves the current status of an intent lease.
//
// Parameters:
//   - intentID (string): The unique identifier for the intent.
//
// Returns:
//   - LeaseStatus: The current status of the lease.
//   - error: An error if the lease is not found.
//
// Errors:
//   - Returns "lease not found for intent" if the lease does not exist.
//
// Side Effects:
//   - None.
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
