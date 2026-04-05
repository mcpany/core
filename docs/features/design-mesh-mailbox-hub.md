# Design Doc: Mesh-Resident Mailbox Hub
**Status:** Draft
**Created:** 2026-04-05

## 1. Context and Scope
As AI agents move from single-threaded task execution to horizontal "Agent Teams" (pioneered by Claude Code), the primary bottleneck has shifted from model reasoning to coordination latency. Current implementations rely on git-based directory locking or unauthenticated local WebSockets, which are either too slow or insecure. The Mesh-Resident Mailbox Hub provides a high-performance, Zero-Trust coordination layer that lives within MCP Any, allowing agents to coordinate at sub-millisecond speeds.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement lock-free teammate coordination using Conflict-Free Replicated Data Types (CRDTs).
    *   Provide hardware-attested task claiming to prevent "Teammate Ghosting".
    *   Support cross-framework synchronization (e.g., a Claude agent delegating to an OpenClaw specialist).
*   **Non-Goals:**
    *   Replacing the agent's internal reasoning loop.
    *   Managing long-term project storage (the hub is for active mission state).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Enable 10 parallel agents to coordinate on a massive refactor without merge conflicts or coordination deadlocks.
*   **The Happy Path (Tasks):**
    1.  Team Lead initializes a "Mission Session" in MCP Any.
    2.  The Hub creates a task-bound mailbox shard.
    3.  Teammate agents join the session and provide hardware-attested identity tokens.
    4.  Lead agent writes tasks to the CRDT task list.
    5.  Teammates claim tasks by updating the CRDT state; the Hub validates the claim against the agent's attested role.
    6.  State is synchronized across all teammates with zero filesystem locks.

## 4. Design & Architecture
*   **System Flow:**
    `[Agent A] <-> [Mailbox Shard (CRDT)] <-> [MCP Any Hub] <-> [Agent B]`
*   **APIs / Interfaces:**
    *   `POST /mesh/session/init`: Creates a new mission-bound coordination shard.
    *   `GET /mesh/mailbox/stream`: gRPC/WebSocket stream for real-time task updates.
    *   `PATCH /mesh/task/{id}/claim`: Atomic task claiming with identity attestation.
*   **Data Storage/State:**
    *   In-memory CRDT graph, periodically checkpointed to the **Agent-Aware Blackboard**.

## 5. Alternatives Considered
*   **Git-based Coordination**: Rejected due to I/O overhead and the risk of "Lock Starvation" in high-frequency swarms.
*   **Centralized SQL Database**: Rejected as it introduces a single point of failure and higher latency compared to memory-resident CRDTs.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** "Auth-before-Coordination" model. Every mesh interaction requires a hardware-attested lineage token.
*   **Observability:** Integrated with the **Service Mesh Topology Monitor** for real-time visualization of task flow.

## 7. Evolutionary Changelog
*   **2026-04-05:** Initial Document Creation.
