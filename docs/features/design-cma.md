# Design Doc: Consensus Mesh Arbiter (CMA)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale horizontally (e.g., Claude Code Agent Teams), the limitation of Conflict-Free Replicated Data Types (CRDTs) becomes apparent in high-contention scenarios. Concurrent state mutations by multiple specialists often lead to "Negotiation Deadlocks" or "Coordination Stalls" where the swarm cannot reach a consistent worldview.

MCP Any needs to move beyond passive state synchronization to active consensus brokering. The Consensus Mesh Arbiter (CMA) provides the authoritative infrastructure for resolving these collisions using Vector Clocks and weighted agent voting, ensuring high-speed coordination for dense meshes.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement decentralized conflict resolution for the Shared KV Store (Blackboard).
    * Support Vector Clock based state tracking to maintain causal order of mutations.
    * Facilitate weighted voting quorums for resolving intent-conflicts.
    * Neutralize "Coordination Stalls" in Agent Teams with 10+ members.
* **Non-Goals:**
    * Replacing CRDTs for low-contention tasks (CMA is for high-stakes/high-contention commits).
    * Managing individual agent reasoning logic (CMA governs the *result* of the reasoning).

## 3. Critical User Journey (CUJ)
* **User Persona:** Autonomous Swarm Orchestrator
* **Primary Goal:** Resolve conflicting database schema update requests from three specialist agents simultaneously.
* **The Happy Path (Tasks):**
    1. Specialist Agents A, B, and C attempt to commit conflicting schema changes to the Blackboard.
    2. CMA intercepts the requests and generates a Vector Clock conflict report.
    3. CMA initiates a "Consensus Vote" among the swarm based on mission-root priority.
    4. The Mission-Root Agent and a Senior Auditor Agent vote for Agent A's path.
    5. CMA commits Agent A's changes and issues "State Re-alignment" signals to Agents B and C.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
    Agents -->|Conflicting Commits| CMA
    CMA -->|Vector Clock Check| VectorEngine
    VectorEngine -->|Conflict Detected| VotingService
    VotingService -->|Quorum Call| MissionRoot
    VotingService -->|Quorum Call| AuditAgents
    MissionRoot -->|Weight 0.6| VoteCollector
    AuditAgents -->|Weight 0.4| VoteCollector
    VoteCollector -->|Winning State| Blackboard
    Blackboard -->|Re-alignment Signal| Agents
    ```
* **APIs / Interfaces:**
    * `POST /v1/consensus/commit`: Submit a state mutation with Vector Clock metadata.
    * `GET /v1/consensus/status`: Query the progress of a pending resolution.
    * `POST /v1/consensus/vote`: Cast a weighted vote on a conflicting state fragment.
* **Data Storage/State:**
    * Utilizes the existing SQLite Blackboard but adds a `vector_clocks` metadata table for causal tracking.

## 5. Alternatives Considered
* **Global Lock Manager:** Rejected due to prohibitive latency (2s+ stalls) in horizontal swarms.
* **Pure CRDT LWW (Last-Write-Wins):** Rejected because it leads to "Intent Erasure" where critical but slightly delayed instructions are discarded.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All votes must be hardware-attested (TPM) to prevent "Sybil Attacks" by rogue subagents.
* **Observability:** CMA will export "Conflict Density" metrics to the UI to highlight coordination bottlenecks.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
