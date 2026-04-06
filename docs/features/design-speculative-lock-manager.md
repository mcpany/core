# Design Doc: Speculative Lock Manager (SLM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms scale horizontally (e.g., Claude Code Agent Teams), the shared teammate mailbox becomes a significant performance bottleneck. Current CRDT-based synchronization methods often lead to 5s+ coordination stalls when multiple agents attempt to claim or modify the same context shard.

The Speculative Lock Manager (SLM) introduces a "fast-path" for shard access. It allows agents to speculatively modify shards while the definitive ownership is resolved in the background. This minimizes "Reasoning Stalls" and ensures that high-density teammate meshes remain responsive during high-contention tasks.

## 2. Goals & Non-Goals
* **Goals:**
    * Reduce inter-teammate coordination latency by 80%.
    * Provide hardware-attested speculative buffers for shard modifications.
    * Ensure deterministic rollback to the last known good state upon lock conflict.
    * Implement "Auth-before-Splice" to prevent unauthorized intent injection.
* **Non-Goals:**
    * Replacing the underlying CRDT mesh (SLM sits on top of it).
    * Providing long-term persistence for speculative states (they are ephemeral until committed).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Resolve a complex code conflict across 10+ teammate agents without encountering "Mailbox Lock" stalls.
* **The Happy Path (Tasks):**
    1. Agent A requests a speculative lock on `shard-001`.
    2. SLM issues a hardware-attested speculative buffer.
    3. Agent A modifies the buffer and streams updates to the mesh.
    4. Background CRDT process verifies ownership.
    5. SLM "promotes" the speculative buffer to the primary Blackboard state.
    6. Other agents ingest the update without ever encountering a "Blocked" state.

## 4. Design & Architecture
* **System Flow:**
    [Agent Request] -> [SLM Speculative Gate] -> [Hardware-Attested Buffer] -> [Background CRDT Verification] -> [Blackboard Commit]
* **APIs / Interfaces:**
    * `POST /v1/locks/speculative/acquire`: Acquire a speculative buffer for a specific shard.
    * `POST /v1/locks/speculative/commit`: Request promotion of a speculative buffer.
    * `ON_CONFLICT`: SLM triggers automated rollback to `mission-root` anchor.
* **Data Storage/State:**
    * Ephemeral memory-mapped buffers (using `memfd_create`) for speculative states.
    * Signed "Intent Tokens" to track buffer parentage.

## 5. Alternatives Considered
* **Strict Optimistic Locking:** Rejected due to the high cost of rollbacks in deep reasoning chains.
* **Global Sequential Locking:** Rejected due to unacceptable latency in swarms larger than 3 agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All speculative buffers are hardware-bound and cryptographically isolated.
* **Observability:** SLM Monitor in the UI provides a real-time heatmap of shard contention and speculative hit rates.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
