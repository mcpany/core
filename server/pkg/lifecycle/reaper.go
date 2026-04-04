package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LeaseStatus represents the current state of a lease.
//
// Summary: Defines the possible states an active or historical lease can be in.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
type LeaseStatus string

const (
	// StatusActive represents an active and valid lease.
	//
	// Summary: Represents an active and valid lease.
	StatusActive LeaseStatus = "ACTIVE"
	// StatusExpired represents a lease that has passed its expiration time.
	//
	// Summary: Represents a lease that has passed its expiration time.
	StatusExpired LeaseStatus = "EXPIRED"
	// StatusPruned represents a lease that has been manually invalidated or cleaned up.
	//
	// Summary: Represents a lease that has been manually invalidated or cleaned up.
	StatusPruned LeaseStatus = "PRUNED"
)

// Lease represents an intent-bound lease for subagents.
//
// Summary: Defines the core lease structure for tracking resource limits and intent boundaries.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
type Lease struct {
	IntentID           string
	SubagentSessionIDs []string
	Expiry             time.Time
	Status             LeaseStatus
	mu                 sync.RWMutex
}

// SubagentReaper monitors and cleans up orphaned or expired subagent leases.
//
// Summary: Reclaims resources tied to inactive subagents to prevent zombie processes.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
type SubagentReaper struct {
	registry map[string]*Lease
	mu       sync.RWMutex
	ticker   *time.Ticker
	quit     chan struct{}
}

// NewSubagentReaper initializes a new Active Subagent Reaper.
// NewSubagentReaper initializes a new subagent reaper instance.
//
// Summary: Creates a new SubagentReaper ready to process active leases.
//
// Parameters:
//   - None.
//
// Returns:
//   - *SubagentReaper: The initialized reaper instance.
//
// Throws/Errors:
//   - None.
func NewSubagentReaper() *SubagentReaper {
	return &SubagentReaper{
		registry: make(map[string]*Lease),
		quit:     make(chan struct{}),
	}
}

// RegisterBranch associates a parallel reasoning branch with a lease.
//
// Summary: Links a sub-branch of reasoning logic to its parent resource lease.
//
// Parameters:
//   - intentID (string): The intent ID.
//   - ttl (time.Duration): Time to live.
//
// Returns:
//   - *Lease: The lease.
//
// Throws/Errors:
//   - None.
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

// RegisterSubagent attaches an active session to a lease.
//
// Summary: Links an agent process ID to the lease controlling its lifecycle.
//
// Parameters:
//   - leaseID (string): The lease controlling the subagent.
//   - subagentSessionID (string): The ID of the session.
//
// Returns:
//   - error: An error if it fails.
//
// Throws/Errors:
//   - error: If the lease is not found.
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

// RecordHeartbeat prolongs the lease for a given subagent.
//
// Summary: Renews the expiration timer for a subagent actively doing work.
//
// Parameters:
//   - intentID (string): The subagent extending its lease.
//   - signature (string): The subagent signature.
//   - extendBy (time.Duration): Extended duration.
//
// Returns:
//   - error: An error if it fails.
//
// Throws/Errors:
//   - error: If the intent is not found.
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
// PruneIntent forcibly cleans up all leases and subagents tied to an intent.
//
// Summary: Immediately revokes all resources scoped to the given logical task.
//
// Parameters:
//   - intentID (string): The logical intent to terminate.
//
// Returns:
//   - error: An error if it fails.
//
// Throws/Errors:
//   - error: If the intent is not found.
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

// Start launches the background sweeping process.
//
// Summary: Initiates the ticker loop to continuously evaluate and collect expired leases.
//
// Parameters:
//   - ctx (context.Context): Context to control the goroutine lifecycle.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
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
// Stop terminates the background sweep.
//
// Summary: Halts the automatic garbage collection of subagent leases.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
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

// GetLeaseStatus fetches the current state of a lease.
//
// Summary: Retrieves the state indicator for an active intent lease.
//
// Parameters:
//   - leaseID (string): The target lease.
//
// Returns:
//   - LeaseStatus: The active status.
//   - error: An error if the lease is unknown.
//
// Throws/Errors:
//   - error: If the lease is not found.
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
