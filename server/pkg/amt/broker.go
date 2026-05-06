// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package amt

import (
	"errors"
	"sync"
)

// Broker manages attested P2P tunnels across the mesh.
//
// Summary: The AMT Broker facilitates hardware-attested, encrypted P2P connections.
type Broker interface {
	// EstablishTunnel initiates a hardware-attested tunnel.
	//
	// Summary: Establishes a secure, attested tunnel to a remote node.
	//
	// Parameters:
	//   - remoteNodeID (string): The identifier of the target node.
	//   - missionToken (string): The hardware-attested mission root token.
	//
	// Returns:
	//   - string: The unique TunnelID.
	//   - error: An error if the connection fails or attestation is invalid.
	//
	// Errors:
	//   - Returns "invalid mission token" if the token is empty.
	//   - Returns "remote node id cannot be empty" if remoteNodeID is empty.
	//
	// Side Effects:
	//   - Registers a new tunnel session in memory.
	EstablishTunnel(remoteNodeID, missionToken string) (string, error)

	// InvokeRemote securely executes a tool over the tunnel.
	//
	// Summary: Proxies a tool call through an established tunnel.
	//
	// Parameters:
	//   - tunnelID (string): The ID of the established tunnel.
	//   - toolCall (string): The serialized tool invocation request.
	//
	// Returns:
	//   - string: The tool execution result.
	//   - error: An error if the tunnel is invalid or execution fails.
	//
	// Errors:
	//   - Returns "tunnel not found" if tunnelID is unknown.
	//
	// Side Effects:
	//   - Sends network traffic over the P2P connection.
	InvokeRemote(tunnelID, toolCall string) (string, error)

	// ResumeTunnel provides fast-path resumption using session-bound trust.
	//
	// Summary: Quickly resumes a previously attested tunnel using a mesh ticket.
	//
	// Parameters:
	//   - meshTicket (string): The session-bound resumption ticket.
	//
	// Returns:
	//   - string: The TunnelID.
	//   - error: An error if the ticket is invalid or expired.
	//
	// Errors:
	//   - Returns "invalid mesh ticket" if meshTicket is empty or unknown.
	//
	// Side Effects:
	//   - Re-activates a tunnel session.
	ResumeTunnel(meshTicket string) (string, error)
}

type amtBrokerImpl struct {
	mu      sync.RWMutex
	tunnels map[string]string // map[tunnelID]remoteNodeID
	tickets map[string]string // map[meshTicket]tunnelID
}

// NewBroker creates a new instance of the AMT Broker.
//
// Summary: Initializes the AMT Broker for managing mesh tunnels.
//
// Parameters:
//   - None.
//
// Returns:
//   - Broker: The initialized Broker interface.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewBroker() Broker {
	return &amtBrokerImpl{
		tunnels: make(map[string]string),
		tickets: make(map[string]string),
	}
}

// EstablishTunnel initiates a hardware-attested tunnel.
//
// Summary: Establishes a secure, attested tunnel to a remote node.
//
// Parameters:
//   - remoteNodeID (string): The identifier of the target node.
//   - missionToken (string): The hardware-attested mission root token.
//
// Returns:
//   - string: The unique TunnelID.
//   - error: An error if the connection fails or attestation is invalid.
//
// Errors:
//   - Returns "invalid mission token" if the token is empty.
//   - Returns "remote node id cannot be empty" if remoteNodeID is empty.
//
// Side Effects:
//   - Registers a new tunnel session in memory.
func (b *amtBrokerImpl) EstablishTunnel(remoteNodeID, missionToken string) (string, error) {
	if remoteNodeID == "" {
		return "", errors.New("remote node id cannot be empty")
	}
	if missionToken == "" {
		return "", errors.New("invalid mission token")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// In a real implementation, we would verify the TPM signature here
    tokenPrefix := missionToken
    if len(missionToken) > 8 {
        tokenPrefix = missionToken[:8]
    }
	tunnelID := "tunnel-" + remoteNodeID + "-" + tokenPrefix // Mock ID generation
	b.tunnels[tunnelID] = remoteNodeID

	// Create a mock mesh ticket for resumption
	ticket := "ticket-" + tunnelID
	b.tickets[ticket] = tunnelID

	return tunnelID, nil
}

// InvokeRemote securely executes a tool over the tunnel.
//
// Summary: Proxies a tool call through an established tunnel.
//
// Parameters:
//   - tunnelID (string): The ID of the established tunnel.
//   - toolCall (string): The serialized tool invocation request.
//
// Returns:
//   - string: The tool execution result.
//   - error: An error if the tunnel is invalid or execution fails.
//
// Errors:
//   - Returns "tunnel not found" if tunnelID is unknown.
//
// Side Effects:
//   - Sends network traffic over the P2P connection.
func (b *amtBrokerImpl) InvokeRemote(tunnelID, toolCall string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if _, exists := b.tunnels[tunnelID]; !exists {
		return "", errors.New("tunnel not found")
	}

	// In a real implementation, we would send the call over the encrypted P2P link
	return "mock-result-for-" + toolCall, nil
}

// ResumeTunnel provides fast-path resumption using session-bound trust.
//
// Summary: Quickly resumes a previously attested tunnel using a mesh ticket.
//
// Parameters:
//   - meshTicket (string): The session-bound resumption ticket.
//
// Returns:
//   - string: The TunnelID.
//   - error: An error if the ticket is invalid or expired.
//
// Errors:
//   - Returns "invalid mesh ticket" if meshTicket is empty or unknown.
//
// Side Effects:
//   - Re-activates a tunnel session.
func (b *amtBrokerImpl) ResumeTunnel(meshTicket string) (string, error) {
	if meshTicket == "" {
		return "", errors.New("invalid mesh ticket")
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	tunnelID, exists := b.tickets[meshTicket]
	if !exists {
		return "", errors.New("invalid mesh ticket")
	}

	return tunnelID, nil
}
