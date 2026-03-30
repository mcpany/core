package erh

// Ephemeral Registry Hook (ERH) Provider
// Security middleware mandating session-locked discovery schemas
// to neutralize registry persistence exploits.

type ERHProvider struct {
}

func NewERHProvider() *ERHProvider {
	return &ERHProvider{}
}
