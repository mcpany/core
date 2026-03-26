# Design Doc: Entangled State Broker (ESB)
**Status:** Draft
**Created:** 2026-06-16

## 1. Context and Scope
As agent swarms move toward high-frequency state sharing via sharded meshes, the risk of "Context Poisoning" and unauthorized state mutation by specialist subagents has reached a critical level. Current "Passive Sanitization" models are insufficient against sub-millisecond MTTC (Mean Time To Compromise). MCP Any needs a proactive mechanism to ensure that state fragments remain cryptographically bound to the mission-root intent, preventing any ingestion of unauthorized state.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-attested "Entanglement Shards" for inter-teammate coordination.
    * Cryptographically bind state fragments to the mission-root intent.
    * Trigger immediate hardware-level corruption signals upon unauthorized mutation.
    * Support sub-100ms state synchronization within sharded meshes.
* **Non-Goals:**
    * Encrypting the entire state for all agents (performance bottleneck).
    * Managing long-term archival of state fragments.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (e.g., Claude Code Team Lead)
* **Primary Goal:** Share high-frequency state fragments between 5 specialized teammates without risking mission-root contamination.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent initializes a new mission scope with the ESB.
    2. The ESB generates hardware-attested "Entanglement Keys" bound to the mission-root TPM.
    3. Teammates request state shards from the ESB, receiving cryptographically entangled fragments.
    4. A subagent attempts to mutate a shard outside its authorized intent branch.
    5. The ESB detects the mutation via hardware-bound integrity checks and triggers a "Shard Corruption" signal.
    6. The parent reasoning engine automatically rolls back the mission branch before re-ingesting the poisoned state.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Init Mission| ESB[Entangled State Broker]
        ESB -->|Generate Keys| TPM[Hardware TPM/Enclave]
        T1[Teammate 1] -->|Request Shard| ESB
        ESB -->|Issue Entangled Shard| T1
        T1 -->|Mutate Shard| ESB
        ESB -->|Verify Lineage| TPM
        TPM -->|Valid| Commit[Commit Mutation]
        TPM -->|Invalid| Alert[Trigger Shard Corruption Signal]
    ```
* **APIs / Interfaces:**
    * `POST /v1/entanglement/init`: Initialize mission-bound entanglement keys.
    * `POST /v1/entanglement/shard/mount`: Request a cryptographically entangled state shard.
    * `POST /v1/entanglement/shard/commit`: Commit a mutation to an entangled shard.
* **Data Storage/State:**
    * Shards are stored in a memory-mapped, zero-copy buffer with hardware-bound integrity tags (MACs).

## 5. Alternatives Considered
* **Full State Encryption:** Rejected due to the prohibitive latency of per-call decryption in high-density meshes.
* **Passive Semantic Scanning:** Rejected as it cannot detect "Low-and-Slow" semantic drift before ingestion.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All entanglement keys are hardware-bound and session-specific. "Shard Corruption" signals are non-maskable.
* **Observability:** Real-time monitoring of "Entanglement Drift" and "Corruption Events" via the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-16:** Initial Document Creation.
