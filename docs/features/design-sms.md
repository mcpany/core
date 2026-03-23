# Design Doc: Sharded Mailbox Sovereignty (SMS)
**Status:** Draft
**Created:** 2026-05-31

## 1. Context and Scope
In deep agent swarms and horizontal agent teams, the inter-agent "Mailbox" (the communication channel for coordination) often becomes a single point of failure and a performance bottleneck. Global mailbox locks or a monolithic coordination bus can't scale to handle hundreds of tasks across multiple agents. Simultaneously, "Mailbox Injection" by rogue subagents remains a critical threat.

Sharded Mailbox Sovereignty (SMS) is an advanced security and performance extension that provides granular, task-bound mailbox shards. This ensures that teammates only have access to the specific communication fragments required for their assigned tasks, anchored to the Mission Root.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement "Task-Bound" mailbox shards to isolate inter-teammate coordination.
    *   Eliminate global mailbox locks for parallel agent teams.
    *   Enforce "Mission-Root" anchoring for every mailbox message.
    *   Provide real-time semantic analysis of state fragments as they cross mailbox boundaries.
*   **Non-Goals:**
    *   Replacing the primary A2A transport (SMS is an isolation and sharding layer *within* the transport).
    *   Storing long-term memory or large blobs (handled by the Shared KV Store).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Horizontal Agent Swarm Orchestrator (e.g., Claude Code Agent Team)
*   **Primary Goal:** Isolate 10 parallel teammates so each can only access mailbox messages related to their specific, assigned task shard.
*   **The Happy Path (Tasks):**
    1.  The Mission Root agent delegates Task A to Agent 1 and Task B to Agent 2.
    2.  The SMS middleware automatically creates two isolated mailbox shards: `shard:mission_root:task_a` and `shard:mission_root:task_b`.
    3.  Agent 1 can only read/write to Shard A, even if both agents share the same parent and mission root.
    4.  The T2T (Teammate-to-Teammate) Encryption Bridge handles the transport, but the SMS layer enforces the shard boundary.
    5.  A "Message Re-composition" audit is performed before Shard A results are shared back to the Mission Root.

## 4. Design & Architecture
*   **System Flow:**
    [Teammate A] <---> [SMS Shard: Task 1] <---> [Mission Root]
    [Teammate B] <---> [SMS Shard: Task 2] <---> [Mission Root]
                          |
                          v
            [Fragment Integrity Monitor (FAMI)]
                          |
                          v
            [Mailbox Integrity Middleware]

*   **APIs / Interfaces:**
    *   `POST /v1/mailbox/shard/create`: Create a task-bound shard anchored to a mission root.
    *   `GET /v1/mailbox/shard/list`: List shards authorized for the current hardware-attested session.
    *   `POST /v1/mailbox/shard/message`: Send a message to a specific shard.
*   **Data Storage/State:**
    *   Ephemeral, task-bound message buffers (in-memory or SQLite per shard).
    *   Cryptographic "Shard Tokens" tied to the Mission Root.

## 5. Alternatives Considered
*   **Monolithic Mailbox with RBAC:** Rejected as it doesn't solve the "Mailbox Lock" performance issue and is prone to "Token Exhaustion" in deep swarms.
*   **Individual Agent-to-Agent Channels:** Rejected as it becomes impossible to audit and maintain "Mission Root" sovereignty across a mesh of 100+ point-to-point connections.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Every SMS shard must be signed with a hardware-attested, session-bound "Mesh Token" (CMCS).
*   **Observability:** Real-time visualization via the "Teammate Task-List Arbiter Workspace" and "Fragment-Aware Mailbox Isolation (FAMI)" auditor.

## 7. Evolutionary Changelog
*   **2026-05-31:** Initial Document Creation.
