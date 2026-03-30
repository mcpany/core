# Design Doc: Horizontal Identity Persistence (HIP) Bridge
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
As agents move from linear, single-framework sessions to horizontal teammate meshes (e.g., Claude Code Agent Teams), the "Session Token" model is failing. Agents that rotate between roles or frameworks lose their mission context and reputation, leading to "Identity Decay" and redundant handshakes.

The HIP Bridge provides hardware-locked, cross-session identity tokens that persist across teammate rotations. This ensures that an agent's "SWARM Reputation" and authorized "Mission Manifest" follow it as it moves between different execution environments within a mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Mesh-Resident" identity root that survives individual agent termination.
    * Provide hardware-attested (TPM) tokens for cross-framework identity mapping.
    * Enable "Fast-Path Identity Resumption" for horizontal teammates.
    * Securely persist identity fragments in the Mesh-Resident Attestation Registry.
* **Non-Goals:**
    * Creating a permanent, global identity for AI agents (focus is on mission/mesh lifetime).
    * Supplanting framework-specific auth (e.g., Anthropic's session tokens); HIP acts as a bridge.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Framework Swarm Orchestrator
* **Primary Goal:** Maintain a consistent security context for a specialist agent that moves from an OpenClaw instance to a Claude Code session.
* **The Happy Path (Tasks):**
    1. The agent completes its task in OpenClaw and initiates a handoff.
    2. The HIP Bridge generates a hardware-locked identity fragment bound to the active Mission Root.
    3. The agent rotates to a Claude Code session and provides its HIP token.
    4. The HIP Bridge validates the token against the Mesh-Resident Registry.
    5. The agent's reputation and mission-bound capabilities are instantly resumed without requiring a full re-handshake with the human user.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent Framework A] -- Handoff --> HIP[HIP Bridge]
        HIP -- Sign --> T[HIP Token]
        T -- Persist --> R[Registry]
        B[Agent Framework B] -- Resume --> HIP
        HIP -- Verify --> R
        Registry -- Attest --> B
    ```
* **APIs / Interfaces:**
    * `POST /v1/hip/checkpoint`: Store identity fragment and sign with TPM.
    * `POST /v1/hip/resume`: Validate fragment and restore capabilities.
    * `X-HIP-Token`: Header for cross-mesh authentication.
* **Data Storage/State:**
    * Identity fragments are stored in a TPM-protected shard within the Mesh-Resident Registry.
    * Monotonic counters are used to prevent replay attacks during resumption.

## 5. Alternatives Considered
* **Short-Lived JWTs:** Rejected because they don't provide hardware-bound lineage and are susceptible to token theft in long-running missions.
* **Manual Re-Attestation:** Rejected because it causes "Coordination Stall" and degrades the performance of autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Identity resumption requires proof of mission continuity and valid hardware attestation.
* **Observability:** Track resumption latency and "Identity Rotation" events in the Horizontal Identity Manager dashboard.

## 7. Evolutionary Changelog
* **2026-07-12:** Initial Document Creation.
