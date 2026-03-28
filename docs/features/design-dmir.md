# Design Doc: Dynamic Mesh Identity Rotation (DMIR)
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
As AI agent swarms scale into high-density meshes, "Identity Squatting" has emerged as a primary attack vector. A compromised subagent or a hijacked session token currently grants an attacker long-term access to sensitive tools. Static session tokens are no longer sufficient for Zero Trust mesh governance.

DMIR evolves the Agent Identity Hub by implementing high-frequency, hardware-bound identity rotation. Every agent interaction is tied to a short-lived token that requires a TPM-signed heartbeat to remain valid, ensuring that identities are temporally sovereign and cryptographically non-reusable.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement sub-millisecond identity rotation for all connected mesh agents.
    * Mandate hardware-attested (TPM/Secure Enclave) heartbeats for session maintenance.
    * Provide a unified "Identity Mint" for cross-framework rotation (Claude, OpenClaw, AutoGen).
    * Support "Mission-Bound Revocation": automatically invalidate all rotated tokens upon mission completion.
* **Non-Goals:**
    * Replacing user-level authentication (OIDC/SAML).
    * Managing tool-specific permissions (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Architect
* **Primary Goal:** Prevent a hijacked "Data Specialist" agent from accessing the Database tool after its assigned 300-second task window expires.
* **The Happy Path (Tasks):**
    1. The Specialist Agent authenticates with the DMIR Provider and receives an initial 60-second token.
    2. The Agent performs a tool call; the DMIR middleware validates the token.
    3. Every 30 seconds, the Agent's local runtime sends a TPM-signed heartbeat to the DMIR Provider.
    4. The DMIR Provider rotates the token, issuing a new one and invalidating the old one.
    5. When the mission completes, the Root Agent signals the DMIR Provider to cease rotation for that lineage.
    6. Any subsequent tool calls from the Specialist using a stale token are blocked.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant DMIR as DMIR Provider
        participant TPM as Hardware Enclave
        participant Tool as MCP Tool
        Agent->>DMIR: Heartbeat + LastToken + TPM Signature
        DMIR->>TPM: Verify Signature
        TPM-->>DMIR: Verified
        DMIR->>Agent: New Rotated Token
        Agent->>Tool: Request + Rotated Token
        Tool->>DMIR: Validate Token
        DMIR-->>Tool: Token Valid
    ```
* **APIs / Interfaces:**
    * `identity.Rotate(heartbeat, signature) -> NewToken`: Rotates the identity fragment.
    * `identity.RevokeLineage(missionRootToken) -> bool`: Ceases rotation for an entire branch.
* **Data Storage/State:**
    * **Active Token Cache:** High-speed in-memory store for verifying rotated fragments with sub-millisecond latency.

## 5. Alternatives Considered
* **Short-lived JWTs without Heartbeats:** Rejected because it doesn't prevent "Squatting" within the token's lifetime if the agent is compromised.
* **mTLS for every request:** Rejected due to the high MTTC (Mean Time to Coordinate) tax in large-scale meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** DMIR is the "Temporal Lock" for the mesh. It ensures that compromise is contained within a 30-60 second window.
* **Observability:** Monitored via the "Dynamic Identity Rotation Dashboard" in the UI.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation. Evolving from the Zero-Trust Agent Identity Hub to support high-frequency temporal sovereignty.
