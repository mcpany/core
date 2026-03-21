# Design Doc: Shared State Arbiter (SSA)
**Status:** Draft
**Created:** 2026-03-21

## 1. Context and Scope
As AI agent swarms move from linear task delegation to horizontal "Agent Teams," the Shared KV Store (Blackboard) has become a primary bottleneck and stability risk. Current implementations suffer from "Reasoning Loops" where multiple specialized subagents attempt to lock the same context shard for atomic refinement, leading to circular wait states and infinite resource consumption.

The Shared State Arbiter (SSA) is a coordination service for the Blackboard that proactively manages state access. It moves the system from a first-come-first-served locking model to an authoritative arbitration model that understands mission-root priority and agent lineage.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time wait-graph analysis to identify and break circular dependencies.
    * Enforce mission-aligned lock prioritization based on agent lineage and task criticality.
    * Provide a standardized "Try-Lock" API with hardware-attested timeouts.
    * Enable atomic "Snapshot-and-Rollback" for shards during conflict resolution.
* **Non-Goals:**
    * SSA will NOT handle the persistence of the KV data itself (handled by the Blackboard).
    * SSA will NOT perform semantic validation of the data content (handled by the AID Hub).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Orchestrator
* **Primary Goal:** Coordinate 5 specialized agents working on a large codebase without deadlocking the shared memory shards.
* **The Happy Path (Tasks):**
    1. Agent A requests a write-lock on `shard:refactor_logic`.
    2. Agent B requests a write-lock on `shard:refactor_logic` while A is holding it.
    3. SSA detects the contention and checks the lineage tokens of both agents.
    4. SSA determines Agent B has a higher mission-priority (e.g., it is a Supervisor agent).
    5. SSA issues a "Checkpoint-and-Yield" signal to Agent A.
    6. Agent A snapshots its internal state to the Blackboard and releases the lock.
    7. Agent B executes its critical update.
    8. SSA re-notifies Agent A that the lock is available for resumption.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> Request Lock (Token) -> SSA (Wait-Graph Check) -> Lock Granted/Queued -> Blackboard Write`
* **APIs / Interfaces:**
    * `POST /v1/ssa/acquire`: Request a timed lock on a specific shard.
    * `POST /v1/ssa/release`: Release a held lock.
    * `GET /v1/ssa/graph`: Retrieve the current wait-graph for observability.
* **Data Storage/State:**
    * SSA maintains an in-memory directed graph of lock owners and waiters.
    * State is cryptographically bound to the Mission Root session.

## 5. Alternatives Considered
* **Timeouts Only**: Rejected because simple timeouts lead to "Thundering Herd" problems and lost reasoning progress when agents are forcefully killed without state recovery.
* **Full Data Locking**: Rejected because locking the entire Blackboard prevents parallel execution in horizontal teams. Granular sharding is required.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All lock requests must be accompanied by a hardware-attested Mission Token. SSA verifies that the requesting agent is authorized to access the specific shard lineage.
* **Observability:** SSA will export "Wait-Time Heatmaps" and "Contention Alerts" to the UI Roadmap items to help developers identify reasoning bottlenecks.

## 7. Evolutionary Changelog
* **2026-03-21:** Initial Document Creation.
