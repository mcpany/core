package interop

import (
	"context"
	"fmt"
)

// UMMB (Universal Multimodal Memory Bus) defines the core interfaces
// for hardware-attested, intent-pinned memory synchronization.

// IntentPinnedStateSerializer normalizes context fragments into a standardized, cryptographically signed format.
type IntentPinnedStateSerializer interface {
	SerializeState(ctx context.Context, shard *MemoryShard) ([]byte, error)
	VerifyLineage(ctx context.Context, shard *MemoryShard) error
}

// MultimodalTraceValidator performs structural and semantic analysis on non-textual inputs.
type MultimodalTraceValidator interface {
	SanitizeTrace(ctx context.Context, payload []byte) ([]byte, error)
}

// HardwareAttestedMemoryShard represents a secure memory region in the Blackboard.
type HardwareAttestedMemoryShard interface {
	ReadShard(ctx context.Context, tpmProof string) (*MemoryShard, error)
	WriteShard(ctx context.Context, shard *MemoryShard, tpmProof string) error
}

// MultimodalHashChainingProvider mandates cryptographic hash-chaining for all non-textual fragments.
type MultimodalHashChainingProvider interface {
	GenerateChainHash(ctx context.Context, currentPayload []byte, previousHash string) (string, error)
	VerifyChain(ctx context.Context, shard *MemoryShard, previousHash string) error
}

// MemoryBusMediator acts as the unified state mediator handling state handoff and validation.
type MemoryBusMediator struct {
	IPSS IntentPinnedStateSerializer
	MTV  MultimodalTraceValidator
	HAMS HardwareAttestedMemoryShard
	MHC  MultimodalHashChainingProvider
}

// PushState handles the state handoff process from an agent framework.
func (m *MemoryBusMediator) PushState(ctx context.Context, shard *MemoryShard, tpmProof string, previousHash string) error {
	if err := m.IPSS.VerifyLineage(ctx, shard); err != nil {
		return fmt.Errorf("lineage verification failed: %w", err)
	}

	if len(shard.MultimodalPayload) > 0 {
		sanitized, err := m.MTV.SanitizeTrace(ctx, shard.MultimodalPayload)
		if err != nil {
			return fmt.Errorf("multimodal sanitization failed: %w", err)
		}
		shard.MultimodalPayload = sanitized

		if m.MHC != nil {
			if err := m.MHC.VerifyChain(ctx, shard, previousHash); err != nil {
				return fmt.Errorf("hash chain verification failed: %w", err)
			}
		}
	}

	if err := m.HAMS.WriteShard(ctx, shard, tpmProof); err != nil {
		return fmt.Errorf("failed to write shard: %w", err)
	}
	return nil
}

// PullState retrieves a sanitized state shard.
func (m *MemoryBusMediator) PullState(ctx context.Context, tpmProof string) (*MemoryShard, error) {
	return m.HAMS.ReadShard(ctx, tpmProof)
}
