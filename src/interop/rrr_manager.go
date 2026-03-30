package interop

import (
	"context"
	"fmt"
)

// RecursiveResourceReclamationManager (RRR) reclaims unused token and reasoning
// budgets from dormant sub-missions.
//
// Intent: Ensures economic efficiency in deep, parallel swarms by reclaiming
// unused token leases.
type RecursiveResourceReclamationManager struct {
	SubMissionBudgets map[string]*ResourceLease
}

type ResourceLease struct {
	SubMissionID string
	TokensLeased int
	TokensUsed   int
	IsDormant    bool
}

// NewRRRManager creates a new RRR Manager instance.
func NewRRRManager() *RecursiveResourceReclamationManager {
	return &RecursiveResourceReclamationManager{
		SubMissionBudgets: make(map[string]*ResourceLease),
	}
}

// AllocateLease allocates a resource lease to a sub-mission.
func (m *RecursiveResourceReclamationManager) AllocateLease(ctx context.Context, subMissionID string, tokens int) {
	m.SubMissionBudgets[subMissionID] = &ResourceLease{
		SubMissionID: subMissionID,
		TokensLeased: tokens,
		TokensUsed:   0,
		IsDormant:    false,
	}
}

// MarkDormant marks a sub-mission as dormant.
func (m *RecursiveResourceReclamationManager) MarkDormant(ctx context.Context, subMissionID string) error {
	lease, exists := m.SubMissionBudgets[subMissionID]
	if !exists {
		return fmt.Errorf("sub-mission %s not found", subMissionID)
	}
	lease.IsDormant = true
	return nil
}

// ReclaimResources forcefully reclaims unused tokens from a dormant sub-mission.
func (m *RecursiveResourceReclamationManager) ReclaimResources(ctx context.Context, subMissionID string) (int, error) {
	lease, exists := m.SubMissionBudgets[subMissionID]
	if !exists {
		return 0, fmt.Errorf("sub-mission %s not found", subMissionID)
	}

	if !lease.IsDormant {
		return 0, fmt.Errorf("sub-mission %s is not dormant, cannot reclaim", subMissionID)
	}

	reclaimed := lease.TokensLeased - lease.TokensUsed
	lease.TokensLeased = lease.TokensUsed // Adjust lease to actual usage

	return reclaimed, nil
}
