# Design Doc: Mesh-Resident Governance Oracles (MRGO)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
As AI agent swarms transition from centralized orchestration to decentralized meshes, the central gateway becomes a critical bottleneck and a single point of failure. Current policy enforcement requires round-tripping to a central authority, introducing latency that degrades swarm performance.

MRGO enables decentralized policy arbitration by hosting "Resident Oracles" within the agent mesh. These oracles provide local, hardware-attested policy decisions, allowing swarms to operate with high autonomy and low latency while maintaining the security guarantees of the Universal Agent Bus.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-millisecond policy arbitration for tool calls within the mesh.
    * Mandate hardware-attested (TPM/TEE) signatures for all local policy decisions.
    * Support seamless synchronization of global security policies to resident oracles.
    * Enable high availability of governance during network partitions.
* **Non-Goals:**
    * Replacing the central policy engine for global auditing and compliance.
    * Providing general-purpose compute for subagents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Decentralized Swarm Orchestrator
* **Primary Goal:** Execute high-frequency tool calls across 10+ subagents without central gateway latency.
* **The Happy Path (Tasks):**
    1. The orchestrator deploys MRGO Adapters alongside subagent clusters.
    2. Central policy is synchronized to MRGO instances and cryptographically locked to local TPMs.
    3. A subagent initiates a tool call.
    4. The local MRGO Adapter intercepts the call and validates it against the resident policy.
    5. MRGO issues a hardware-signed approval token locally.
    6. The tool executes immediately, and the decision is asynchronously reported to the central audit log.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent] -->|Tool Call| B[MRGO Adapter]
        B -->|Local Policy Check| C{Policy Engine}
        C -->|Approved| D[TPM Signer]
        D -->|Hardware Token| E[Tool Provider]
        E -->|Execution Result| A
        B -.->|Async Report| F[Central Audit Log]
        G[Central Policy Hub] -.->|Sync| B
    ```
* **APIs / Interfaces:**
    * `rpc Arbitrate(ToolCallRequest) returns (HardwareAttestedToken)`
    * `rpc SyncPolicy(SignedPolicyUpdate) returns (Ack)`
* **Data Storage/State:**
    * Local policies are stored in a TPM-sealed SQLite database within the MRGO sidecar.

## 5. Alternatives Considered
* **Centralized Gateways:** Rejected due to high latency (100ms+) and lack of partition resilience.
* **Pure Agent-Side Enforcement:** Rejected because it lacks hardware-bound isolation and can be bypassed if the agent is compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All MRGO decisions must be signed by a hardware root of trust. Tokens are time-bound and mission-scoped.
* **Observability:** decision traces are streamed to the central hub via a non-blocking telemetry channel.

## 7. Evolutionary Changelog
* **2026-06-18:** Initial Document Creation.
