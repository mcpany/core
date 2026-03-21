# Design Doc: Lock-Free Mesh Arbiter (LFMA)
**Status:** Draft
**Created:** 2026-05-31

## 1. Context and Scope
As AI agent swarms evolve from hierarchical trees to horizontal teams (e.g., Claude Code's Agent Teams), the traditional model of a single "Supervisor" or "Mailbox Lock" is becoming a performance bottleneck. Parallel teammates attempting to claim tasks or synchronize state frequently collide, leading to "Cognitive Stall" and increased latency.

The Lock-Free Mesh Arbiter (LFMA) is designed to provide a decentralized, non-blocking coordination layer for horizontal swarms. It utilizes Conflict-Free Replicated Data Types (CRDTs) to ensure that teammates can claim, delegate, and update tasks asynchronously without requiring global locks or a central supervisor for every state change.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide sub-millisecond task claiming for parallel teammates.
    *   Ensure eventual consistency of the shared task list across framework-neutral meshes.
    *   Support lock-free state synchronization for horizontal swarms exceeding 5+ agents.
    *   Integrate with hardware-attested identity to ensure secure, non-repudiable coordination.
*   **Non-Goals:**
    *   Replacing the mission-root intent (LFMA manages task state, not mission definition).
    *   Providing real-time synchronous consensus for every minor tool call (handled by CQ Hub).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Horizontal Agent Swarm Orchestrator (e.g., Claude Code Agent Team)
*   **Primary Goal:** Enable 10 parallel teammates to claim and execute tasks from a shared list without coordination locks.
*   **The Happy Path (Tasks):**
    1.  The Mission Root agent initializes a shared task list on the LFMA bus.
    2.  Parallel Teammates (Claude, OpenClaw, AutoGen) subscribe to the LFMA mesh.
    3.  Agent A identifies a task and issues a "Claim" instruction via a CRDT LWW-Element-Set.
    4.  Agent B simultaneously attempts to claim the same task; the LFMA resolves the conflict locally using monotonic timestamps and hardware-attested identity priority.
    5.  Both agents continue reasoning without waiting for a global lock acknowledgment.
    6.  The mesh state converges asynchronously, and Agent B gracefully pivots to the next available task.

## 4. Design & Architecture
*   **System Flow:**
    Teammate A (Claim) ---> [LFMA CRDT Buffer] <--- Teammate B (Claim)
                                |
                                v
                    [Conflict Resolution Logic]
                                |
                                v
                    [Hardware-Attested Mesh State]
                                |
                                v
                    Broadcast Update to all Teammates

*   **APIs / Interfaces:**
    *   `POST /v1/mesh/claim`: Submit a task claim with hardware-attested session token.
    *   `GET /v1/mesh/state`: Retrieve the current converged CRDT state of the task mesh.
    *   `WS /v1/mesh/sync`: Real-time WebSocket stream for sharded state updates.
*   **Data Storage/State:**
    *   In-memory CRDT (G-Set or LWW-Element-Set) for active tasks.
    *   Persistent backup in the Shared KV Store (Blackboard) with STL (State-Trust Labeling).

## 5. Alternatives Considered
*   **Centralized Redis Lock:** Rejected due to the "Local Sovereignty" requirement and the latency of network-bound locking in air-gapped or high-frequency local environments.
*   **Hierarchical Supervisor handoffs:** Rejected as it recreates the supervisor bottleneck we are trying to eliminate for horizontal teams.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Every LFMA operation must be signed with a HAIR (Hardware-Attested Identity Rotation) token to prevent "Teammate Impersonation" or "Task Squatting."
*   **Observability:** Real-time visualization via the "Teammate Task-List Arbiter Workspace" in the UI, showing CRDT merge events and conflict resolutions.

## 7. Evolutionary Changelog
*   **2026-05-31:** Initial Document Creation.
