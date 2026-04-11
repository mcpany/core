# Design Doc: Lock-Free Task Auction (LFTA) Hub
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale horizontally (e.g., Claude Code Agent Teams), the overhead of coordination has become a primary performance bottleneck. Currently, teams often rely on git-based filesystem locks or centralized database transactions to "claim" tasks from a shared list. This leads to "Cognitive Stall" cycles exceeding 5 seconds in high-density teams (10+ agents) due to filesystem I/O and merge conflict resolution.

The Lock-Free Task Auction (LFTA) Hub is designed to solve this by providing a mesh-resident, sub-millisecond synchronization layer using memory-mapped Conflict-Free Replicated Data Types (CRDTs). MCP Any will act as the authoritative host for these task queues, allowing parallel teammates to coordinate without global locks.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-millisecond task claiming and resolution for horizontal agent teams.
    * Use CRDT-based synchronization to eliminate the need for global coordination locks.
    * Utilize memory-mapped files (`memfd_create`) for zero-copy state sharing between teammates.
    * Enforce hardware-attested mission-root authority for all task auctions.
* **Non-Goals:**
    * Replacing long-term task persistence (which remains in SQLite/Git).
    * Managing the internal reasoning of agents; it only manages the coordination fragments.
    * Providing a general-purpose message broker for non-agentic applications.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Coordinate 12 parallel specialist agents on a complex codebase migration without triggering coordination timeouts.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent submits 50 sub-tasks to the LFTA Hub.
    2. 12 specialist agents concurrently query the LFTA Hub for available tasks.
    3. Agents "bid" or "claim" tasks by updating their local CRDT shard in memory-mapped space.
    4. The LFTA Hub automatically merges shard updates, resolving collisions without blocking.
    5. Agents receive near-instant confirmation of task ownership.
    6. As tasks are resolved, the Hub propagates the "Resolved" state to the rest of the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Submit Tasks| B[LFTA Hub]
        B <--> C[Shared Memory CRDT Shard]
        D[Agent 1] <-->|Lock-Free Claim| C
        E[Agent 2] <-->|Lock-Free Claim| C
        F[Agent N] <-->|Lock-Free Claim| C
        C -->|Snapshot & Merge| G[Blackboard / SQLite]
    ```
* **APIs / Interfaces:**
    * `lfta.SubmitAuction(taskBatch, missionToken) -> AuctionID`: Initializes a new task auction.
    * `lfta.ClaimFragment(auctionID, taskID, agentIdentity) -> Ack`: Non-blocking claim using CRDT semantics.
    * `lfta.CommitResolution(auctionID, taskID, result) -> Ack`: Finalizes a task and clears the memory shard.
* **Data Storage/State:**
    * **Primary State:** Memory-mapped `memfd` segments containing task vectors.
    * **Consistency Model:** Strong Eventual Consistency via Automerge or Yjs-compatible CRDT engines.
    * **Persistence:** Periodic asynchronous snapshots to the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Git-Based Locking:** Rejected due to 1-5s latency per operation and high filesystem stress.
* **Centralized Redis/Postgres Locking:** Rejected due to network round-trip overhead and the "Thundering Herd" problem during concurrent claims.
* **Lock-Free shared-memory Atomicity:** Rejected because raw atomics don't handle the complex "Identity + Status" state required for task claiming as well as CRDTs do.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All CRDT operations must be signed with a hardware-attested session token. Unauthorized writes to the `memfd` segment trigger a hardware-level corruption signal.
* **Observability:** Monitored via the "CRDT Shard Monitor" and "Lock-Free Coordination Debugger" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
