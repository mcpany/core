package interop

import (
	"context"
	"fmt"
)

// HardwareAttestedCostAttributionProvider (HACA) cryptographically attributes
// token usage to specific sub-process lineage.
//
// Intent: Eliminates "Economic Squatting" where subagents bleed parent budgets.
type HardwareAttestedCostAttributionProvider struct {
	AttributionLedger map[string]*CostAttribution
}

type CostAttribution struct {
	SubProcessID string
	TokensUsed   int
	GPUSeconds   float64
	Signature    string
}

// NewHACAProvider creates a new HACA Provider instance.
func NewHACAProvider() *HardwareAttestedCostAttributionProvider {
	return &HardwareAttestedCostAttributionProvider{
		AttributionLedger: make(map[string]*CostAttribution),
	}
}

// AttributeCost records hardware-attested cost usage for a specific sub-process.
func (p *HardwareAttestedCostAttributionProvider) AttributeCost(ctx context.Context, subProcessID string, tokens int, gpu float64, signature string) error {
	if signature == "" {
		return fmt.Errorf("cryptographic signature required for cost attribution")
	}

	p.AttributionLedger[subProcessID] = &CostAttribution{
		SubProcessID: subProcessID,
		TokensUsed:   tokens,
		GPUSeconds:   gpu,
		Signature:    signature,
	}
	return nil
}

// GetAttribution retrieves the verified cost attribution for a sub-process.
func (p *HardwareAttestedCostAttributionProvider) GetAttribution(subProcessID string) (*CostAttribution, error) {
	attr, exists := p.AttributionLedger[subProcessID]
	if !exists {
		return nil, fmt.Errorf("no attribution found for sub-process %s", subProcessID)
	}
	return attr, nil
}
