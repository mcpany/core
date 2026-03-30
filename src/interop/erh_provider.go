package interop

import (
	"context"
	"fmt"
	"time"
)

// EphemeralRegistryHookProvider (ERH) mandates session-locked discovery schemas.
//
// Intent: Neutralizes "Registry Persistence" exploits where malicious subagents
// shadow legitimate tools in long-running sessions.
type EphemeralRegistryHookProvider struct {
	ActiveHooks map[string]*EphemeralHook
}

type EphemeralHook struct {
	HookID    string
	Schema    string
	ExpiresAt time.Time
}

// NewERHProvider creates a new ERH Provider instance.
func NewERHProvider() *EphemeralRegistryHookProvider {
	return &EphemeralRegistryHookProvider{
		ActiveHooks: make(map[string]*EphemeralHook),
	}
}

// IssueSessionLockedHook issues a time-bound, session-locked hook.
func (p *EphemeralRegistryHookProvider) IssueSessionLockedHook(ctx context.Context, schema string, duration time.Duration) string {
	hookID := fmt.Sprintf("erh-%d", time.Now().UnixNano())
	p.ActiveHooks[hookID] = &EphemeralHook{
		HookID:    hookID,
		Schema:    schema,
		ExpiresAt: time.Now().Add(duration),
	}
	return hookID
}

// RetrieveHook retrieves an ephemeral hook if it has not expired.
func (p *EphemeralRegistryHookProvider) RetrieveHook(ctx context.Context, hookID string) (string, error) {
	hook, exists := p.ActiveHooks[hookID]
	if !exists {
		return "", fmt.Errorf("hook %s not found", hookID)
	}

	if time.Now().After(hook.ExpiresAt) {
		delete(p.ActiveHooks, hookID)
		return "", fmt.Errorf("hook %s has expired", hookID)
	}

	return hook.Schema, nil
}
