# Design Doc: DRC (Deterministic Reasoning Checkpoint) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Deep agent swarms often encounter "Cognitive Stalls" or "Reasoning Deadlocks" when a specialist subagent diverges from the mission root or hallucinates a recursive dependency. Currently, the only recourse is a full session restart, which is expensive in terms of token usage and time.

The DRC Hub is required to implement the OpenClaw v3.6.2 standard, allowing agents to create hardware-attested checkpoints of their cognitive state (Internal Monologue + Blackboard State) before attempting high-risk or complex operations. This enables "Reasoning Rewinding" to a known-good fragment without mission-root loss.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate TPM-signed snapshots of agent monologues and shared blackboard states.
    * Enable sub-millisecond rewinding to specific reasoning fragments upon detection of drift or failure.
    * Support "Speculative Checkpointing" for parallel teammate branches.
    * Neutralize cognitive stalls by providing a deterministic state-recovery mechanism.
* **Non-Goals:**
    * Archiving the entire conversation history (that is the job of the logger).
    * Providing manual "undo" for users (this is an agent-to-agent recovery utility).
    * Managing non-deterministic model weights.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Recover a complex coding mission after a "Refactoring Specialist" agent incorrectly deletes a core module due to a hallucinatory loop.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a refactoring mission and instructs the DRC Hub to create a "Pre-Refactor" checkpoint.
    2. DRC Hub takes a TPM-signed snapshot of the current Blackboard state and the parent reasoning trace.
    3. The specialist subagent begins the refactor; the Agentic Entropy Monitor (AEM) detects an immediate spike in reasoning entropy.
    4. The AEM signals a "Cognitive Divergence" to the parent agent.
    5. The parent agent invokes the DRC Hub to rewind to the "Pre-Refactor" checkpoint.
    6. DRC Hub restores the Blackboard and reasoning context in <10ms.
    7. The parent agent prunes the specialist subagent and re-delegates the task to a different teammate.

## 4. Design & Architecture
* **System Flow:**
    [Agent] --(Checkpoint Signal)--> [DRC Hub] --(Snapshot)--> [Hardware Vault]
    [AEM Monitor] --(Divergence Detected)--> [Agent]
    [Agent] --(Rewind Signal)--> [DRC Hub] --(Restore)--> [Blackboard/Monologue]
* **APIs / Interfaces:**
    * `drc.CreateCheckpoint(missionID, tag) -> CheckpointID`: Captures current state with TPM signature.
    * `drc.RewindToCheckpoint(checkpointID) -> RestoredState`: Performs the atomic rewind.
    * `drc.PruneCheckpoints(missionID/beforeTime)`: Lifecycle management for snapshots.
* **Data Storage/State:**
    * **Hardware Vault:** TPM-encrypted storage for state fragments.
    * **Lineage Graph:** Metadata mapping checkpoints to specific reasoning branches.

## 5. Alternatives Considered
* **Standard SQLite Savepoints:** Rejected because they lack hardware-bound attestation. A compromised agent could manipulate or delete its own savepoints.
* **Full Session Logging:** Rejected because restoration from raw logs is too slow and non-deterministic for real-time mesh coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Checkpoints are immutable once signed and can only be restored by the mission-root or authorized supervisor.
* **Observability:** Integrated with the "Swarm Rollback Dashboard" for visual tracking of rewind events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
