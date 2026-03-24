# Design Doc: Lock-Free Mesh Arbiter (LFMA)
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
As AI agent teams scale horizontally (e.g., Claude Code swarms), traditional centralized task queues and database locks become a prohibitive bottleneck. The "Mailbox Lock" crisis of early 2026 saw coordination latency spike to 2s+ in swarms of 10+ agents.

The Lock-Free Mesh Arbiter (LFMA) implements a decentralized coordination model using Conflict-Free Replicated Data Types (CRDTs) to provide sub-millisecond task synchronization across heterogeneous framework boundaries.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Eliminate coordination deadlocks in high-density horizontal swarms.
    *   Achieve <10ms task-claiming latency for teammate mailboxes.
    *   Ensure mission-root consistency across decentralized framework nodes.
    *   Provide hardware-attested conflict resolution for parallel agent edits.
*   **Non-Goals:**
    *   Replacing the Shared KV Store (Blackboard) for persistent storage.
    *   Implementing a full distributed database (LFMA is for coordination state).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Coordinate 20 parallel specialist agents on a complex refactoring task without mailbox contention.
*   **The Happy Path (Tasks):**
    1.  The parent agent publishes a list of 50 tasks to the sharded mailbox.
    2.  Specialist agents utilize the LFMA client to perform local, lock-free task claiming.
    3.  LFMA synchronizes the CRDT task list across all teammate nodes in the background.
    4.  Conflict resolution is handled via hardware-attested logical clocks (LWW-Element-Set).
    5.  Teammates stream status updates to the shared shard, visible to the parent without blocking.

## 4. Design & Architecture
*   **System Flow:**
    [Parent Agent] -> (Seed Tasks) -> [CRDT Mailbox Shard]
                                            |
                    -------------------------------------------------
                    |                       |                       |
            [Teammate A (OpenClaw)]   [Teammate B (Claude)]   [Teammate C (AutoGen)]
                    |                       |                       |
            (Claim Task 1)           (Claim Task 2)          (Claim Task 3)
                    |                       |                       |
                    ------- (P2P Sync / CRDT Merge) -----------------

*   **APIs / Interfaces:**
    *   `POST /v1/mesh/claim`: Atomic, lock-free task claiming.
    *   `GET /v1/mesh/state`: Retrieve current CRDT state for a mailbox shard.
    *   `POST /v1/mesh/sync`: Peer-to-peer state synchronization endpoint.
*   **Data Storage/State:**
    *   In-memory LWW-Element-Set (Last-Write-Wins) for task ownership.
    *   Merkle-Search-Trees (MST) for efficient delta synchronization.

## 5. Alternatives Considered
*   **Centralized Redis Queue:** Rejected due to single-point-of-failure and latency in local-first deployments.
*   **Raft-based Consensus:** Rejected as too heavy for the high-frequency, transient state of agent mailboxes.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All CRDT operations must be signed with hardware-bound identity tokens to prevent "State Splicing."
*   **Observability:** Visualized via the "Lock-Free Coordination Monitor," showing real-time shard throughput and conflict rates.

## 7. Evolutionary Changelog
*   **2026-06-28:** Initial Document Creation.
