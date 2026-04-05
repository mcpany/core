# Design Doc: Speculative Reflection Arbiter (SRA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Speculative Reflection Loops (SRL) in Claude Code and similar patterns in other agent frameworks, agents are now performing actions before a global consensus or mission-root reflection is complete. This reduces "Cognitive Stall" but introduces "Reflection Poisoning," where un-validated speculative state pollutes the shared Blackboard, leading to swarm divergence.

MCP Any needs to provide a secure mediation layer that allows agents to benefit from speculative speed while maintaining a deterministic "Return to Truth" upon quorum failure.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide isolated "Probabilistic Buffers" for speculative tool results.
    * Implement atomic commit/rollback based on Mission-Root reflection signals.
    * Support framework-agnostic speculative signaling (UACO v3.8 compliance).
* **Non-Goals:**
    * Predicting the outcome of the reflection quorum.
    * Replacing the framework's internal reflection logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Enable speculative agent speed without risking Blackboard corruption.
* **The Happy Path (Tasks):**
    1. Agent A initiates a high-stakes tool call with a `x-speculative-intent` header.
    2. SRA intercepts the call and routes the output to a volatile `Probabilistic Shard`.
    3. Specialist teammates reason against the `Probabilistic Shard` in an isolated branch.
    4. Mission-Root Quorum completes reflection and issues a `MISSION_COMMIT` signal.
    5. SRA atomically merges the `Probabilistic Shard` into the global Blackboard.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [Speculative Header] -> MCP Any (SRA) -> Volatile Buffer -> [Commit Signal] -> Global Blackboard`
* **APIs / Interfaces:**
    * `POST /v1/state/speculative/prepare`: Initialize a speculative session.
    * `POST /v1/state/speculative/commit`: Promote buffer to persistent state.
    * `POST /v1/state/speculative/abort`: Purge buffer and dependent branches.
* **Data Storage/State:**
    * Utilizes in-memory LRU cache for `Probabilistic Shards` to ensure zero-latency access before commit.

## 5. Alternatives Considered
* **Client-Side Rollback**: Rejected because it relies on agent honesty and doesn't protect the shared bus from other teammates ingesting poisoned state.
* **Synchronous-Only Execution**: Rejected due to the 5s+ "Cognitive Stall" bottleneck identified in current swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative state is marked as `Low-Trust` and cannot be used to trigger non-speculative tool calls (e.g. `rm -rf /`).
* **Observability:** Visualized in the `Speculative State Inspector` dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
