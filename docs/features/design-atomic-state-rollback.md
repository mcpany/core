# Design Doc: Atomic State Rollback (ASR) Middleware
**Status:** Draft
**Created:** 2026-03-29

## 1. Context and Scope
As agent swarms become more autonomous and perform complex multi-step reasoning, the risk of "Swarm Divergence" or "State Poisoning" by a single specialized agent increases. The Atomic State Rollback (ASR) middleware allows parent agents to create swarm-wide "Checkpoints" of the collective state (Blackboard, Context Shards, and Session Metadata). If a sub-task fails or deviates from the mission intent, the entire environment can be reverted to a known-good checkpoint.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a `CheckpointManager` that can snapshot the state of the Blackboard and active Context Shards.
    * Provide an atomic `Rollback(checkpointID)` operation that reverts all state mutation since the checkpoint.
    * Support "Temporal Snapshots" in the Virtual Context Map.
    * Integrate with the UACO v1.9 MAQ for consensus-based rollback triggers.
* **Non-Goals:**
    * Providing long-term archival of checkpoints (they are session-bound and ephemeral).
    * Automatically detecting when a rollback is needed (this is triggered by the Parent Agent or a Monitor).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Coordinator Agent
* **Primary Goal:** Recover from a failed code refactoring attempt by a specialized "Refactor-Subagent" without manual cleanup.
* **The Happy Path (Tasks):**
    1. Parent Agent initiates a `Checkpoint` before delegating the refactoring task.
    2. Subagent performs multiple tool calls, mutating the Blackboard and code shards.
    3. Parent Agent (or a Monitor) detects that the refactoring has introduced a regression.
    4. Parent Agent triggers a `Rollback` using the `checkpointID`.
    5. ASR Middleware reverts all Blackboard entries and Shard states to the exact moment the checkpoint was taken.

## 4. Design & Architecture
* **System Flow:**
    `Task Start` -> `Create Checkpoint` -> `Mutations...` -> `Anomaly Detected` -> `Trigger Rollback` -> `State Restored`
* **APIs / Interfaces:**
    * `CheckpointManager`: `CreateCheckpoint(sessionID string) (id string, err error)`, `Rollback(id string) error`, `Discard(id string) error`
* **Data Storage/State:**
    * Checkpoints are stored as copy-on-write (CoW) deltas in the Shared KV Store and BSH State Buffer to minimize memory overhead.

## 5. Alternatives Considered
* **Sequential Tool Undo**: Rejected as many tools (e.g., shell commands) are not easily undoable. ASR provides environmental atomicity.
* **Full State Cloning**: Rejected due to high performance and memory costs in deep swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Rollback triggers must be authorized by the Parent Agent or a verified MAQ Quorum.
* **Observability:** Checkpoint creation and rollback events are logged to the "Swarm Rollback Dashboard."

## 7. Evolutionary Changelog
* **2026-03-29:** Initial Document Creation.
* **2026-03-30: IPSC-Triggered Checkpoints**
    * **Context:** The discovery of "Cognitive Lock" in self-correction loops (UACO v2.1) requires more granular recovery options.
    * **Architecture Adjustment:** Added "Auto-Checkpointing" on the first correction cycle of an IPSC session. If the Correction Budget is exceeded, the ASR middleware can now offer an "Atomic Rollback to Pre-Correction State."
    * **Security Impact:** Prevents "State Smearing" where failed refinements leave residue in the Shared KV Store.
