# Design Doc: Attested Mesh Tunneling (AMT)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms move from single-machine environments to distributed meshes (OpenClaw Sovereign Node Tunneling), the "Universal Agent Bus" must provide secure, low-latency interconnectivity across physical device boundaries. Current inter-node coordination relies on "Implicit Local Trust" or heavy VPN overhead, which both fail to meet the performance and security demands of sub-millisecond agentic execution.

AMT provides a native, hardware-attested P2P tunneling layer for MCP Any, ensuring that inter-node tool calls and state synchronization are encrypted and origin-locked without the latency tax of traditional network overlays.

## 2. Goals & Non-Goals
* **Goals:**
    * Establish hardware-attested (TPM/Secure Enclave) P2P encrypted tunnels between MCP Any nodes.
    * Provide sub-millisecond tunnel resumption using "Mesh Tickets."
    * Enforce origin-locked sovereignty across node boundaries (preventing browser-bridge hijacking).
    * Seamlessly integrate with the A2A Messaging Hub for cross-node task delegation.
* **Non-Goals:**
    * Replace general-purpose VPNs for non-agentic traffic.
    * Provide public internet discovery (AMT focuses on peered/authenticated nodes).

## 3. Critical User Journey (CUJ)
* **User Persona:** Prosumer with a Multi-Node Agent Mesh (e.g., Desktop + Edge Home Server).
* **Primary Goal:** Execute a high-privilege tool on a home server from a laptop agent without exposing the server to the local network.
* **The Happy Path (Tasks):**
    1. Laptop Agent (Node A) discovers a tool capability on Home Server (Node B) via the A2A Discovery Bus.
    2. Node A initiates an AMT handshake with Node B, providing a hardware-attested identity token.
    3. Node B verifies the token against its local trust-root and establishes an encrypted P2P tunnel.
    4. Node A executes the tool call over the tunnel; Node B enforces local Zero-Trust policies.
    5. Future calls utilize a "Mesh Ticket" for sub-millisecond resumption.

## 4. Design & Architecture
* **System Flow:**
    [Agent Node A] <--> [AMT Broker] <== Encrypted P2P Tunnel ==> [AMT Broker] <--> [Agent Node B]
    - Handshake: Noise Protocol Framework + TPM Signatures.
    - Transport: QUIC (UDP) for low-latency stream multiplexing.
* **APIs / Interfaces:**
    - `POST /v1/mesh/handshake`: Initiate P2P connection with attestation payload.
    - `POST /v1/mesh/ticket`: Exchange session-bound resumption tickets.
* **Data Storage/State:**
    - Encrypted "Mesh Tickets" stored in secure kernel-bound memory.
    - Peer trust-roots managed via the Distributed Mesh Identity Hub.

## 5. Alternatives Considered
* **WireGuard/Tailscale Integration**: Rejected due to the inability to natively bind network-level trust to reasoning-level hardware attestation (TPM) without external orchestration.
* **Legacy TLS/mTLS**: Rejected due to high handshake latency (100ms+) for ephemeral agentic coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All mesh traffic is subject to "Auth-at-the-Pipe" validation. Even within a tunnel, every tool call requires an active mission-root token and is subject to SES-compliant metadata sanitization.
* **Observability:** AMT connection health and latency metrics are exported to the Service Mesh Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
