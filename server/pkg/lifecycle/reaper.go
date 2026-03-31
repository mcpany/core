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
	// StatusActive represents an active and valid lease.
	StatusActive LeaseStatus = "ACTIVE"
	// StatusExpired represents a lease that has passed its expiration time.
	StatusExpired LeaseStatus = "EXPIRED"
	// StatusPruned represents a lease that has been manually invalidated or cleaned up.
	StatusPruned LeaseStatus = "PRUNED"
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
func NewSubagentReaper() *SubagentReaper {
	return &SubagentReaper{
		registry: make(map[string]*Lease),
		quit:     make(chan struct{}),
	}
}

// RegisterBranch creates a new speculative intent branch and assigns a lease.
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
