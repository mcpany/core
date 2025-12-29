// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package balancer

import (
	"sync/atomic"
)

// Balancer defines the interface for selecting a target from a list of addresses.
type Balancer interface {
	// Next returns the next address to use.
	Next() string
}

// RoundRobinBalancer implements the Balancer interface using a round-robin strategy.
type RoundRobinBalancer struct {
	addresses []string
	current   uint64
}

// NewRoundRobinBalancer creates a new RoundRobinBalancer.
func NewRoundRobinBalancer(addresses []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		addresses: addresses,
		current:   0,
	}
}

// Next returns the next address in the rotation.
func (b *RoundRobinBalancer) Next() string {
	if len(b.addresses) == 0 {
		return ""
	}
	if len(b.addresses) == 1 {
		return b.addresses[0]
	}

	next := atomic.AddUint64(&b.current, 1)
	return b.addresses[(next-1)%uint64(len(b.addresses))]
}
