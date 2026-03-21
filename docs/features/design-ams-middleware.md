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
