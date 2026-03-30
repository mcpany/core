package zksa

// Zero-Knowledge State Attestation (ZKSA) Provider
// Elevated to support DMR-compliant state migration proofs.

type ZKSAProvider struct {
}

func NewZKSAProvider() *ZKSAProvider {
	return &ZKSAProvider{}
}
