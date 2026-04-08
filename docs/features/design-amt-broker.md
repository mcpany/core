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
        end
        subgraph Node B (Workstation)
            D[AMT Broker] --> E[Local MCP Server]
            D --> F[TPM Verification]
        end
        B <==>|Attested P2P Tunnel| D
    ```
* **APIs / Interfaces:**
    * `amt.EstablishTunnel(remoteNodeID, missionToken) -> TunnelID`: Initiates a hardware-attested tunnel.
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
### Update: [2026-07-25] - Neutralizing RCE and Sequential Tool Attack Chains
**Context:** July 2026 security research into OpenClaw (arXiv:2603.12644) has identified that standard mesh tunnels are vulnerable to **Sequential Tool Attack Chains**, where a compromised subagent uses valid P2P tunnels to bridge from unprivileged to privileged tool contexts.
**Architecture Adjustment:**
- **Reasoning-Action Correlation**: Integrating the `Sequential Attack-Chain Interceptor` into the AMT handshake. Tunnels are no longer granted based on static hardware identity alone; they now require a **Lineage-Aware Intent Proof** that correlates the remote tool call with the cumulative reasoning trace of the mission-root.
- **Dynamic Context Redaction**: AMT Brokers will now perform real-time semantic scrubbing of context fragments crossing the tunnel to prevent "Context Amnesia" exploits used in RCE-driven prompt injection.
**Security Impact:** Mitigates the risk of a remote node being used as a lateral movement jump-box by a rogue parallel teammate.
