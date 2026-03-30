package erh

import (
	"context"
)

// Provider is the Ephemeral Registry Hook (ERH) Provider
// Security middleware mandating session-locked discovery schemas
// to neutralize registry persistence exploits.
type Provider struct {
}

// NewProvider creates a new ERH Provider
func NewProvider() *Provider {
	return &Provider{}
}

// MandateSessionLocked mandates session-locked discovery schemas
func (p *Provider) MandateSessionLocked(ctx context.Context, schema string) error {
	// Placeholder implementation
	return nil
}
