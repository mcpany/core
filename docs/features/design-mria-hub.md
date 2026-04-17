# Design Doc: Mesh-Resident Identity Attestation (MRIA) Hub
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
With the rise of "Sovereign Node Tunneling" (OpenClaw) and horizontal "Agent Teams" (Claude Code), agentic systems are moving from single-machine environments to distributed meshes. Traditional "Local Trust" models are failing because browser-based attackers or rogue local processes can hijack unauthenticated loopback/named-pipe traffic. MCP Any needs to provide a hardware-attested identity that is sovereign across these tunnels and multi-node handoffs.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) identity tokens to local and remote agents.
    * Ensure identity persistence across P2P tunnels (e.g., OpenClaw SNT).
    * Neutralize "Implicit Local Trust" by mandating cryptographic handshakes for all inter-node coordination.
    * Support sub-millisecond identity verification for high-frequency coordination.
* **Non-Goals:**
    * Replacing global OAuth providers (MRIA is for mesh-local sovereignty).
    * Managing non-agentic user identities.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Securely delegate a task from a local Claude instance to a remote OpenClaw specialist on another device without exposing local env vars or trusting the network.
* **The Happy Path (Tasks):**
    1. The Primary Agent requests an MRIA token from the local MCP Any Hub.
    2. MCP Any Hub verifies the local environment (TPM check) and issues a session-bound, mission-anchored identity token.
    3. The Primary Agent initiates an SNT tunnel to the Remote Node.
    4. The Remote Node's MCP Any Hub mandates an MRIA handshake before exposing any capabilities.
    5. The Primary Agent provides the hardware-attested token; the Remote Hub validates the lineage back to the Mission Root.
    6. Coordination proceeds over the encrypted, authenticated tunnel.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[Agent] -->|Request Identity| Hub[MRIA Hub]
        Hub -->|TPM/Enclave Check| HW[Hardware Root]
        Hub -->|Issue Signed Token| Agent
        Agent -->|Handshake| RemoteHub[Remote MRIA Hub]
        RemoteHub -->|Verify Lineage| Hub
    ```
* **APIs / Interfaces:**
    * `POST /identity/mint`: Returns a hardware-signed JWT containing environment and mission metadata.
    * `POST /identity/verify`: Validates a token and returns the trust level and lineage.
* **Data Storage/State:**
    * Ephemeral, session-bound registry of active mission-root tokens.
    * In-memory cache of verified remote hub public keys.

## 5. Alternatives Considered
* **Static mTLS**: Rejected due to the overhead of certificate management in highly dynamic, ephemeral agent swarms.
* **Shared Secret (API Keys)**: Rejected as they are easily exfiltrated and do not provide environment attestation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All inter-node coordination is denied by default until a valid MRIA handshake is completed. Tokens are cryptographically bound to the mission-root to prevent lateral movement.
* **Observability:** Centralized audit log of all identity minting and cross-node handshakes, providing a "Chain of Command" trace.

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation.
