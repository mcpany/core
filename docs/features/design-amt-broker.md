# Design Doc: Attested Mesh Tunneling (AMT) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms evolve from single-node instances to distributed meshes (e.g., OpenClaw SNT), the "Implicit Local Trust" assumption has become a catastrophic vulnerability. Malicious subagents on the same local network can probe and hijack unencrypted inter-agent coordination channels. MCP Any needs to provide a secure, authenticated, and encrypted transport layer that works across physical device boundaries while maintaining absolute origin-locked sovereignty.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested, encrypted P2P tunnels for multi-node agent meshes.
    * Mandate cryptographic handshakes for all inter-node tool calls and state handoffs.
    * Maintain origin-locked sovereignty across physical boundaries.
    * Support sub-millisecond tunnel resumption for high-frequency coordination.
* **Non-Goals:**
    * Implementation of a generic VPN service (scope is restricted to agentic traffic).
    * Providing public internet exposure for local tools (tunnels are mesh-internal).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Node Swarm Orchestrator
* **Primary Goal:** Securely delegate a high-privilege filesystem task from a mobile agent to a desktop worker node without exposing the coordination to the local network.
* **The Happy Path (Tasks):**
    1. The mobile agent (Mission Root) initiates a task delegation to the desktop worker.
    2. MCP Any (AMT Broker) on the mobile node intercepts the request and identifies the remote destination.
    3. The AMT Broker initiates a hardware-attested handshake with the desktop node's AMT listener.
    4. Both nodes verify each other's TPM/Secure Enclave signatures and mission-root lineage.
    5. A session-bound, encrypted P2P tunnel is established.
    6. The task delegation and subsequent tool results are exchanged securely over the tunnel.
    7. The tunnel is automatically torn down or moved to a "Fast-Path" resume state upon task completion.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A] <-> [AMT Broker A] <-Encrypted P2P Tunnel (TPM-Signed)-> [AMT Broker B] <-> [Agent B]`
* **APIs / Interfaces:**
    * `POST /mesh/tunnel/init`: Initiates a hardware-attested handshake with a remote node.
    * `WS /mesh/tunnel/stream`: The encrypted transport channel for A2A and tool traffic.
    * `GET /mesh/nodes/attest`: Retrieves the hardware-attestation status of known mesh peers.
* **Data Storage/State:**
    * Node identities and mission-root lineages are stored in the hardware-bound `Mesh Identity Store`.
    * Tunnel session keys are rotated periodically and never persist to disk.

## 5. Alternatives Considered
* **Standard WireGuard/mTLS:** Rejected because they lack native "Mission-Root" and "Reasoning-Path" awareness. A standard VPN can't verify if a subagent is diverging from its parent's intent.
* **Legacy HTTP Proxies:** Rejected due to lack of encryption and vulnerability to MitM attacks on the local network.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tunnels are only established after successful hardware-attestation of both nodes. All traffic is origin-locked.
* **Observability:** Tunnel establishment, latency, and attestation failures are logged to the `Mesh Audit Sink`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
