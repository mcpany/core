# Design Doc: Consensus Integrity Monitor

**Status:** Draft
**Created:** 2026-05-14

## 1. Context and Scope
As agent swarms evolve toward "Massively Parallel Execution" (e.g., Claude Code Agent Teams, OpenClaw swarms), maintaining a consistent worldview across parallel reasoning branches becomes a critical infrastructure requirement. "Consensus Fragmentation" occurs when subagents diverge in their state, leading to conflicting tool calls or mission failure. The Consensus Integrity Monitor acts as the authoritative arbiter to ensure all parallel branches remain aligned with a single, non-fragmentable consensus.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a high-speed reconciliation service for merging parallel agent reasoning states.
    * Detect "Reasoning Hijacking" where one agent coerces another via shared context.
    * Implement "Consensus Barriers" to ensure all branches reach alignment before executing high-risk tools.
    * Maintain a cryptographically signed "Consensus Root" for the entire swarm.
* **Non-Goals:**
    * Automatically resolving every minor semantic difference (some divergence is expected for exploration).
    * Replacing the primary agent's task-specific logic.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Ensure that three parallel subagents working on a large codebase maintain a consistent understanding of the project's dependency graph.
* **The Happy Path (Tasks):**
    1. The parent agent spawns 3 parallel subagents.
    2. Subagent A discovers a new circular dependency.
    3. The Consensus Integrity Monitor intercepts A's state update and broadcasts a "Consensus Barrier."
    4. Subagents B and C are paused until they acknowledge and ingest the new dependency state.
    5. The Monitor verifies "Consensus Alignment" and releases the barrier.
    6. All subagents proceed with a unified worldview.

## 4. Design & Architecture
* **System Flow:**
    * Integrated with the **Parallel Team Coordination Hub**.
    * Uses a "Snapshot-and-Merge" model where subagents commit state fragments to the Monitor.
    * The Monitor performs "Conflict Detection" using semantic similarity and rule-based integrity checks.
* **APIs / Interfaces:**
    * `CommitState(branch_id, state_fragment)`: Subagents push new reasoning results.
    * `AcknowledgeBarrier(branch_id, consensus_id)`: Subagents confirm alignment with the merged state.
    * `VerifyConsensus(swarm_id)`: Returns the current authoritative consensus for the swarm.
* **Data Storage/State:** Uses the **Shared KV Store (Blackboard)** for persistent state and a high-speed memory-mapped buffer for active barrier management.

## 5. Alternatives Considered
* **Peer-to-Peer Synchronization**: Rejected due to high latency and the risk of "Byzantine" failures in deep swarms.
* **Parent-Only Arbitration**: Rejected because the parent agent becomes a performance bottleneck and cannot monitor sub-monologues in real-time.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Consensus merges require multi-signature attestation from the participating agents.
* **Observability**: "Consensus Drift" metrics and "Reasoning Hijacking" alerts are surfaced in the Swarm Health Dashboard.

## 7. Evolutionary Changelog
* **2026-05-14:** Initial Document Creation.
