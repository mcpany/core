# Design Doc: Asynchronous Mailbox Sharding (AMS) Middleware
**Status:** Draft
**Created:** 2026-03-20

## 1. Context and Scope
With the rise of "Agent Teams" (e.g., Claude Code, OpenClaw swarms), inter-agent coordination has become a major performance bottleneck. Current implementations often rely on a single, global "Blackboard" or synchronous "Mailbox Locks" where only one teammate can update state or claim a task at a time. This leads to "Mailbox Contention" and significant latency in high-density teammate swarms.

The AMS Middleware solves this by sharding the inter-agent mailbox based on the mission task list. Instead of a monolithic queue, AMS hosts granular, task-bound mailbox shards. This allows parallel teammates to synchronize state and claim sub-tasks without global coordination locks, enabling horizontal scalability for the teammate mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement granular, task-bound mailbox shards.
    * Provide lock-free state synchronization for parallel teammates.
    * Ensure mission-root intent is pinned to every shard for security alignment.
    * Support "Ghost Task" reclamation via heartbeat monitoring.
* **Non-Goals:**
    * Implementing the transport layer itself (handled by Named Pipes/WebSockets).
    * Defining the agent reasoning logic (handled by the LLM).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 10 specialized teammates working on a complex codebase refactor without "Mailbox Lock" bottlenecks.
* **The Happy Path (Tasks):**
    1. The mission root agent initializes the AMS Hub with a sharded task list.
    2. 10 teammates connect to MCP Any and subscribe to their respective shards.
    3. Teammates A and B simultaneously claim "Refactor Module X" and "Refactor Module Y" from different shards.
    4. AMS processes both claims in parallel without wait-states.
    5. Teammate C updates the "Module X" shard with refactored code metadata.
    6. Teammate A receives the update instantly via the sharded event stream.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MissionRoot[Mission Root Agent] -->|Init Shards| AMSHub[AMS Middleware]
        AMSHub --> Shard1[Task Shard: Refactor]
        AMSHub --> Shard2[Task Shard: Test]
        AMSHub --> Shard3[Task Shard: Audit]

        TeammateA[Teammate A] <-->|Claim/Update| Shard1
        TeammateB[Teammate B] <-->|Claim/Update| Shard2
        TeammateC[Teammate C] <-->|Claim/Update| Shard3

        Shard1 -.->|Sync| Blackboard[Shared KV Store]
        Shard2 -.->|Sync| Blackboard
        Shard3 -.->|Sync| Blackboard
    ```
* **APIs / Interfaces:**
    * `POST /ams/shard/create`: Creates a new mailbox shard for a task ID.
    * `PUT /ams/shard/claim`: Atomically claims a task fragment within a shard.
    * `GET /ams/shard/stream`: WebSocket endpoint for shard-level event streaming.
* **Data Storage/State:**
    * Shard states are held in high-speed, in-memory buffers backed by the SQLite Blackboard for persistence.

## 5. Alternatives Considered
* **Global Table Locking:** Rejected due to unacceptable latency in swarms larger than 3 agents.
* **Pure CRDTs:** Rejected for the primary mailbox because "Task Claiming" requires atomic consensus (one-and-only-one teammate per task). AMS uses CRDTs for state sync but atomic locks for fragment claiming.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every shard request must be signed with an IFA (Identity Fragment Attestation) token to prevent teammate impersonation.
* **Observability:** AMS provides a "Mailbox Shard Monitor" UI to visualize task allocation and contention hotspots.

## 7. Evolutionary Changelog
* **2026-03-20:** Initial Document Creation.

### Update: 2026-03-22 - Lock-Free Mesh Coordination
**Context:** Today's market sync reveals that horizontal Agent Teams are hitting a performance ceiling due to synchronous "Mailbox Locks" when synchronizing shared task lists.
**Architecture Adjustment:**
* Deprecating global synchronization locks in Section 4.
* Introducing **Lock-Free Coordination** utilizing Conflict-Free Replicated Data Types (CRDTs) for shard-level state synchronization.
* Implementing atomic "Task Claiming" via optimistic concurrency control (OCC) to ensure strict single-agent assignment without global wait-states.
**Security Impact:** Reduces the risk of "Coordinated DoS" where a single lagging teammate stalls the entire mesh coordination bus.

### Update: 2026-06-27 - CRDT-Native Mailbox Sharding
**Context:** As teammate meshes scale beyond 10+ parallel agents, even optimistic locks on shards are causing "Coordination Stall" (2s+ latency).
**Architecture Adjustment:**
* **CRDT-Native Shards**: Transitioning from OCC to full CRDT-native mailbox shards (OR-Set with LWW-Register per task).
* **State Streaming**: Utilizing the BSH (Binary State Handoff) gateway to stream delta-CRDT updates instead of full-state syncs.
* **Deterministic Tie-Breaking**: Hardware-attested identity priority is now the primary tie-breaker for concurrent task claims, eliminating the need for any back-and-forth lock negotiation.
**Security Impact:** Prevents "Mailbox Splicing" (CVE-2026-81042) by mandating that every CRDT mutation be signed with an IFA token.
