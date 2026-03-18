# Design Doc: Federated Swarm Identity (FSI) Provider
**Status:** Draft
**Created:** 2026-05-23

## 1. Context and Scope
In the current ecosystem of "Heterogeneous Swarms," AI agents from different frameworks (Claude Code, OpenClaw, AutoGen) must collaborate to solve complex tasks. However, there is no universal identity standard, leading to "Identity Fragmentation" where agents rely on weak, framework-specific proxy IDs. This makes swarms vulnerable to identity spoofing and unauthorized task delegation.

FSI provides a local, hardware-attested "Identity Mint" that issues cross-framework identity tokens. These tokens allow any agent in the mesh to verify the lineage, framework origin, and mission-bound authority of its peers.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Issue hardware-attested (TPM/Secure Enclave) identity tokens to local and remote agents.
    *   Provide a framework-agnostic "Identity Registry" where agents can look up peer capabilities.
    *   Enable "Lineage Validation": verify that a subagent was spawned by an authorized parent.
    *   Support cryptographically bound "Agent Cards" for secure discovery.
*   **Non-Goals:**
    *   Replacing global identity providers (e.g., Auth0, Google). FSI is for agent-to-agent mesh identity.
    *   Managing user identities. FSI focuses on NHI (Non-Human Identities).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Systems Architect orchestrating a multi-framework security audit swarm.
*   **Primary Goal:** Ensure that a specialized "Vulnerability Scanner" agent (OpenClaw) only accepts tasks from the verified "Lead Orchestrator" (Claude Code) and can prove its own identity to the "Report Writer" (AutoGen).
*   **The Happy Path (Tasks):**
    1.  The Lead Orchestrator registers with the FSI Provider and receives a hardware-bound Master Token.
    2.  The Lead spawns a Scanner subagent and requests a "Delegate Token" from FSI, bound to the parent's lineage.
    3.  The Scanner receives the token and uses it to authenticate its RPC calls to other teammates.
    4.  Teammates verify the Scanner's identity and lineage against the FSI Registry.
    5.  FSI logs the identity handshake in the secure audit trail.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent as AI Agent (Claude/OpenClaw)
        participant FSI as FSI Provider
        participant TPM as Secure Enclave (TPM)
        participant Peer as Peer Agent
        Agent->>FSI: Request Identity Token (Framework Metadata)
        FSI->>TPM: Sign Identity Claims
        TPM-->>FSI: Hardware-Attested Signature
        FSI->>Agent: FSI Token (JWT with TPM Signature)
        Agent->>Peer: Task Delegation (with FSI Token)
        Peer->>FSI: Verify Token & Lineage
        FSI-->>Peer: Validation Success (Identity/Lineage Info)
    ```
*   **APIs / Interfaces:**
    *   `POST /identity/mint`: Issues a new identity token. Requires framework attestation.
    *   `POST /identity/verify`: Validates a token and returns the agent's lineage/authority.
    *   `GET /identity/registry`: Searchable bus for discovering active peer "Agent Cards."
*   **Data Storage/State:**
    *   Internal SQLite store for tracking active tokens and parent-child lineage mappings.
    *   Secure integration with host TPM for signing.

## 5. Alternatives Considered
*   **Framework-Specific Tokens:** Rejected because they don't solve the cross-framework coordination problem.
*   **Shared API Keys:** Rejected as insecure for mesh coordination (prone to theft and exfiltration).

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** FSI is the foundation of "Auth-before-Discovery." It ensures no "Shadow Agents" can join the mission.
*   **Observability:** Integrated with the `NHI Identity Wallet Status` UI for real-time monitoring of agent identities.

## 7. Evolutionary Changelog
*   **2026-05-23:** Initial Document Creation.
