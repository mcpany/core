# Design Doc: Conflict-Free Sovereignty Broker (CFSB)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms transition from hierarchical, synchronous coordination to horizontal, asynchronous meshes, the reliance on traditional locking mechanisms has led to "Cognitive Stall"—significant latency where specialist agents wait for state availability. The recent move toward Conflict-Free Replicated Data Types (CRDTs) in frameworks like Claude Code has reduced these stalls but introduced "Merge Divergence," where teammates work on conflicting versions of a mission-critical state before reconciliation.

The Conflict-Free Sovereignty Broker (CFSB) is required to act as the authoritative reconciliation hub for MCP Any. It moves beyond basic CRDT merging by applying hardware-attested semantic rules to resolve conflicts, ensuring that the "Winning Intent" always aligns with the mission-root authority.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide authoritative reconciliation for asynchronous state shards and task lists.
    * Implement hardware-attested semantic merging to resolve CRDT divergence.
    * Ensure that all state mutations are anchored to the mission-root priority rules.
    * Neutralize "Merge Divergence" by enforcing monotonic state consistency.
* **Non-Goals:**
    * Replacing the underlying CRDT implementations (CFSB sits on top of them).
    * Resolving conflicts in raw data without mission-root context (all merging must be intent-aware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Horizontal Swarm Orchestrator
* **Primary Goal:** Reconcile conflicting file edits from two parallel specialist agents without a global coordination lock.
* **The Happy Path (Tasks):**
    1. Two specialist agents (A and B) optimistically edit a shared "SharedConfig" shard in their local CRDT mailboxes.
    2. Agent A's update reaches the CFSB first; it is validated against the mission-root manifest.
    3. Agent B's conflicting update arrives; the CFSB detects the divergence.
    4. CFSB performs a "Semantic Merit" check, comparing both updates against the hardware-attested mission intent.
    5. CFSB issues a "Reconciliation Token" that merges the changes, giving priority to the branch with the highest attestation strength or user-defined priority.
    6. Both agents synchronize their local shards to the CFSB-verified state.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A State] \                                  / [Reconciled State]`
    `                   -> [CFSB (Semantic Arbiter)] -> `
    `[Agent B State] /                                  \ [Attestation Token]`
* **APIs / Interfaces:**
    * `/v1/cfsb/push`: Agents submit optimistic state updates.
    * `/v1/cfsb/pull`: Agents fetch the reconciled, mission-root state.
    * `/v1/cfsb/resolve`: Manual or automated trigger for conflict resolution policies.
* **Data Storage/State:**
    * Conflict logs and semantic embeddings are stored in a transient "Reconciliation Buffer" before being committed to the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Last-Writer-Wins (LWW)**: Rejected as it ignores the semantic value of the agent's work and can lead to mission-critical instruction loss.
* **Synchronous Global Locks**: Rejected due to the 5s+ "Cognitive Stall" observed in high-density swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All resolution decisions must be hardware-attested. A compromised subagent cannot "win" a conflict by spoofing priority metadata.
* **Observability:** "Merge Divergence" events and reconciliation latency are tracked in the "Mesh State Debugger."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
