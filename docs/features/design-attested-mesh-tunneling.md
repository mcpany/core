# Design Doc: Attested Mesh Tunneling (AMT)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms evolve from local processes to distributed meshes spanning multiple devices and cloud providers, "Implicit Local Trust" has become a critical vulnerability. Current loopback and VPC-bound communications are susceptible to cross-site hijacking and "Mesh Shadowing" where unauthorized nodes probe agent capabilities.

Attested Mesh Tunneling (AMT) is designed to provide a secure, hardware-attested transport layer for all inter-agent communication. It ensures that every tool call, state transfer, and task delegation is encrypted and non-repudiable across physical and network boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Establish encrypted P2P tunnels between distributed agent nodes using hardware-bound (TPM/SEP) identities.
    * Mandate a cryptographic handshake for every inter-node connection, even on verified local networks.
    * Neutralize "Mesh Shadowing" by making agent capabilities invisible to un-attested nodes.
    * Support sub-millisecond tunnel resumption using session-bound "Mesh Tickets."
* **Non-Goals:**
    * Replacing existing VPN/SD-WAN infrastructure for general network traffic.
    * Providing anonymization services for agent traffic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Securely invoke a local database tool from a specialist agent running on a remote cloud instance without exposing the tool to the open internet.
* **The Happy Path (Tasks):**
    1. The cloud agent requests a connection to the local node.
    2. MCP Any on both nodes performs a hardware-attested mutual TLS (mTLS) handshake.
    3. The AMT Broker establishes an encrypted tunnel bound to the mission-root session.
    4. The cloud agent invokes the database tool through the tunnel.
    5. The local node verifies the cryptographically signed intent before execution.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A (Cloud)] -> [AMT Client (Cloud)] -> (Encrypted P2P Tunnel) -> [AMT Server (Local)] -> [Target Tool]`
* **APIs / Interfaces:**
    * `EstablishTunnel(PeerIdentity, MissionToken) -> TunnelHandle`
    * `RotateMeshTicket(TunnelHandle) -> MeshTicket`
* **Data Storage/State:**
    * Tunnel metadata and "Mesh Tickets" are stored in kernel-bound memory, isolated from subagent process environments.

## 5. Alternatives Considered
* **Standard VPC Peering:** Rejected because it lacks per-agent hardware attestation and granularity.
* **Legacy VPNs:** Rejected due to high latency overhead and complexity of mission-bound lifecycle management.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All tunnels are ephemeral and tied to a specific mission root; compromise of one node does not expose the entire mesh.
* **Observability:** Integrated with the Service Mesh Topology Monitor for real-time visualization of tunnel health and attestation events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
