# Design Doc: Autonomous Service Mesh Gateway
**Status:** Draft
**Created:** 2026-07-09

## 1. Context and Scope
Current multi-agent coordination often relies on unauthenticated local network ports or insecure message passing, creating a massive attack surface for "Teammate Impersonation" and "Context Splicing." As swarms move toward horizontal collaboration (e.g., Claude Code Agent Teams), the "Universal Agent Bus" must move beyond simple bridging to active Mesh Governance.

The Autonomous Service Mesh Gateway provides secure, authenticated transport and discovery for all inter-agent communication. It ensures that agent capabilities are only visible to authorized peers and that all state handoffs are cryptographically bound to a verified mission root.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide authenticated transport for inter-agent communication (mTLS/Named Pipes).
    * Implement "Auth-before-Discovery" for agent capability cards.
    * Mediate lock-free state synchronization between parallel teammates.
    * Enforce mesh-resident security policies (MRPS integration).
* **Non-Goals:**
    * Managing the internal reasoning state of individual agents.
    * Replacing framework-specific coordination logic (e.g., AutoGen's GroupChat).

## 3. Critical User Journey (CUJ)
* **User Persona:** Heterogeneous Swarm Architect
* **Primary Goal:** Securely synchronize a "Shared Task List" between Claude-led managers and OpenClaw-led specialists.
* **The Happy Path (Tasks):**
    1. Agents authenticate with the Service Mesh Gateway using tokens from the Identity Hub.
    2. The Gateway establishes a secure, encrypted channel (T2T Bridge).
    3. The "Shared Task List" is sharded across the mesh, managed by the Gateway's lock-free arbiter (LFMA).
    4. Agents "claim" and "delegate" tasks via hardware-attested handshakes.
    5. The Gateway continuously validates that all mesh traffic aligns with the hardware-attested mission root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] -->|Auth Handshake| G[Service Mesh Gateway]
        B[Agent B] -->|Auth Handshake| G
        G -->|Authenticated Discovery| R[Capability Registry]
        A -->|Encrypted Task Delegation| G
        G -->|Lineage Validation| B
        G <-->|CRDT State Sync| M[Mailbox Shards]
    ```
* **APIs / Interfaces:**
    * `POST /v1/mesh/handshake`: Establishes an authenticated session.
    * `GET /v1/mesh/discovery`: Returns authenticated capability cards.
    * `POST /v1/mesh/coordinate`: Sends mailbox messages or state updates.
* **Data Storage/State:**
    * Uses CRDTs for shared state to ensure non-blocking performance.
    * Persists mesh topology in the Namespace-Locked Discovery service.

## 5. Alternatives Considered
* **Centralized Global Lock:** Rejected due to the 2s+ coordination stall observed in high-density teams.
* **Unauthenticated Peer-to-Peer:** Rejected due to "Teammate Hijacking" risks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandates session-bound authentication for every inter-agent packet.
* **Observability:** Provides a real-time topology view of active mesh coordination and discovery paths.

## 7. Evolutionary Changelog
* **2026-07-09:** Initial Document Creation.
