# Design Doc: Atomic State Rollback (ASR) Middleware
**Status:** Draft
**Created:** 2026-03-29

## 1. Context and Scope
In deep agent swarms, a single specialized subagent's failure or hallucination can poison the shared state (Blackboard) and context sharding, leading to a "cascading hallucination" across the entire system. Following OpenClaw's v2026.3.28 release, MCP Any must provide a mechanism for parent agents to "checkpoint" the collective state and perform atomic rollbacks if a sub-trajectory is deemed invalid.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement `ASRMiddleware` to manage swarm-wide state checkpoints.
    * Provide a `Rollback` API that restores the Blackboard and Context Shards to a specific Consensus Epoch.
    * Integrate with the UACO v2.0 Logical Clock for deterministic state recovery.
    * Support "Nested Checkpoints" for hierarchical agent chains.
* **Non-Goals:**
    * Automating the decision to rollback (this is handled by parent agents or monitor agents).
    * Restoring external state (e.g., real-world file deletions or API side-effects). ASR is for internal agentic state only.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator Agent
* **Primary Goal:** Explore a high-risk reasoning path using a "Refinement" subagent, with the ability to discard all state changes if the refinement fails validation.
* **The Happy Path (Tasks):**
    1. Parent Agent creates a "State Checkpoint" via MCP Any before delegating to Subagent B.
    2. Subagent B performs multiple tool calls, writing to the Blackboard and creating new Context Shards.
    3. Parent Agent (or a Monitor Agent) detects a hallucination in Subagent B's output.
    4. Parent Agent issues an `ASR.Rollback` command.
    5. MCP Any restores the Blackboard and Context Shards to the exact state they were in at Step 1, effectively "erasing" Subagent B's toxic influence on the shared state.

## 4. Design & Architecture
* **System Flow:**
    `Checkpoint Request` -> `State Snapshot (Blackboard + Shards)` -> `Logical Clock Sync` -> `Execution` -> `(Optional) Rollback Request` -> `State Restoration`
* **APIs / Interfaces:**
    * `ASRManager`: `CreateCheckpoint(intentID string) (epochID int64, error)`, `Rollback(epochID int64) error`, `Commit(epochID int64) error`
* **Data Storage/State:**
    * Snapshots are stored as "Copy-on-Write" (CoW) layers in the SQLite Blackboard and the Shard-Aware State Buffer.
    * Metadata about checkpoints (Epoch, Intent, Parentage) is maintained in an immutable "Traceability Log."

## 5. Alternatives Considered
* **Individual Tool Rollbacks**: Rejected as they don't capture the relational state between different tools and context shards.
* **Memory-Only Snapshots**: Rejected as they don't survive agent restarts or session handoffs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Rollback commands must be signed by the Parent Agent who created the checkpoint or a higher-level "Swarm Governor."
* **Observability:** Rollback events and "State Divergence" alerts are visualized in the "Swarm Rollback Dashboard."

## 7. Evolutionary Changelog
* **2026-03-29:** Initial Document Creation.
