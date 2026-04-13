# Design Doc: Attested Mesh Tunneling (AMT) Broker
**Status:** Draft
**Created:** 2026-07-24

## 1. Context and Scope
As AI agents move from single-device environments to distributed multi-node meshes (e.g., OpenClaw's Sovereign Node Tunneling), the risk of "Mesh Shadowing" and unauthenticated inter-node execution has become critical. Standard network-layer VPNs or tunnels are insufficient because they lack the "Agentic Awareness" needed to verify that a remote tool call is bound to a specific mission root and authorized hardware identity.

The Attested Mesh Tunneling (AMT) Broker is required to provide hardware-attested, agent-aware encrypted P2P tunnels that maintain origin-locked sovereignty across physical device boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate hardware-attested, encrypted P2P connections between distributed MCP Any nodes.
    * Enforce mission-bound authorization for all remote tool calls.
    * Neutralize "Mesh Shadowing" by requiring cryptographic handshakes for every inter-node connection.
    * Support "Lightweight Mesh Handshakes" for sub-millisecond tunnel resumption.
* **Non-Goals:**
    * Providing a general-purpose VPN for non-agent traffic.
    * Managing low-level network routing outside the agent bus.
    * Replacing existing local Zero-Trust (LOWA) protocols; it extends them to the mesh.

## 3. Critical User Journey (CUJ)
* **User Persona:** Distributed Swarm Architect
* **Primary Goal:** Securely allow a coding agent on a Laptop to invoke a GPU-accelerated testing tool on a remote Workstation.
* **The Happy Path (Tasks):**
    1. Laptop agent initiates a remote tool call to the Workstation node.
    2. AMT Broker on the Laptop intercepts the request and generates a hardware-attested "Mesh Handshake" token.
    3. The Broker establishes an encrypted P2P tunnel to the Workstation's AMT Broker.
    4. The Workstation Broker verifies the handshake token against the Laptop's TPM signature and mission-root intent.
    5. Once verified, the Workstation Broker proxies the tool call to the local MCP server.
    6. Tool results are returned through the attested tunnel to the Laptop agent.
    7. For subsequent calls, the agents use a session-bound "Mesh Ticket" for fast-path resumption.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        subgraph Node A (Laptop)
            A[Agent] --> B[AMT Broker]
            B --> C[TPM Attestation]
            G[Speculative Prefetcher] -.->|Pre-warm| B
        end
        subgraph Node B (Workstation)
            D[AMT Broker] --> E[Local MCP Server]
            D --> F[TPM Verification]
        end
        B <==>|Attested P2P Tunnel| D
    ```
* **Speculative Tunnel Prefetcher:** A background service that monitors the agent's internal monologue and intent fragments. Upon detecting a high-probability remote tool requirement, it pre-emptively triggers a hardware-attested handshake to establish a P2P tunnel, reducing MTTC.
* **APIs / Interfaces:**
    * `amt.EstablishTunnel(remoteNodeID, missionToken, lineageProof) -> TunnelID`: Initiates a hardware-attested tunnel with mandatory lineage signing.
    * `amt.InvokeRemote(tunnelID, toolCall) -> Result`: Securely executes a tool over the tunnel.
    * `amt.ResumeTunnel(meshTicket) -> TunnelID`: Fast-path resumption using session-bound trust.
* **Data Storage/State:**
    * **Mesh Identity Registry:** Distributed ledger or local cache of verified node hardware fingerprints.
    * **Tunnel Session Store:** In-memory tracking of active tunnels and their mission-bound authorization.

## 5. Alternatives Considered
* **Standard WireGuard/Tailscale Tunnels:** Rejected because they provide network-layer access without agentic mission-binding. A compromised subagent could use the tunnel for unauthorized lateral movement.
* **HTTPS Proxying with API Keys:** Rejected because API keys are susceptible to exfiltration. AMT requires hardware-locked (TPM) attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory origin validation (SOP) is enforced at both ends of the tunnel. Hardware-locked keys prevent tunnel hijacking.
* **Observability:** Integrated with the "Service Mesh Topology Monitor" in the UI for real-time visualization of inter-node tunnels.

## 7. Evolutionary Changelog
* **2026-07-24:** Initial Document Creation.

### Update: 2026-07-25 - Neutralizing Tunneling Latency via Speculative Resumption
**Context:** Today's market sync revealed that mandatory P2P tunnels in OpenClaw are introducing significant latency, impacting sub-millisecond tool execution. Simultaneously, the "Lease-Shadowing" exploit pattern (CVE-2026-10293) requires more granular lineage validation during tunnel establishment.
**Architecture Adjustment:**
* Speculative Tunnel Prefetching: Introducing a background pre-warming service in Section 4 that utilizes real-time intent analysis to establish attested tunnels *before* the remote tool call is issued.
* Lineage-Bound Handshake (LBH): Upgrading the `amt.EstablishTunnel` interface to mandate a cryptographically signed "Lineage Proof," ensuring that the remote node can verify the complete parentage of the subagent before granting tunnel access.
**Security Impact:** Prevents "Lease-Shadowing" where a subagent might inherit parent hardware leases without explicit mission-root re-attestation, while maintaining high performance.
