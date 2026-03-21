# Design Doc: Shared State Arbiter (SSA)
**Status:** Draft
**Created:** 2026-03-21

## 1. Context and Scope
As agent swarms grow in depth and parallelism, multiple agents often attempt to read and write to the same shared memory (the Blackboard) simultaneously. This leads to "State Deadlocks," where agents enter infinite reasoning loops because they cannot agree on the current state of a task. The SSA provides the necessary lock-management and conflict-resolution logic to ensure swarm stability.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a lock-management API for Blackboard keys.
    * Implement version-conflict resolution for parallel state mutations.
    * Support "Intent-Bound" state isolation.
* **Non-Goals:**
    * Replacing the underlying Blackboard storage (SQLite).
    * Providing long-term archival of all state changes (only active session state).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Coordinate 3 specialist agents writing to a shared "Task List" without data corruption or reasoning loops.
* **The Happy Path (Tasks):**
    1. Agent A requests a write-lock for `task_list/1`.
    2. SSA grants the lock and issues a version token.
    3. Agent B requests a write-lock for `task_list/1` and is placed in a prioritized wait-queue.
    4. Agent A commits the update with the version token.
    5. SSA releases the lock and notifies Agent B.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> SSA (Lock Request) -> Blackboard (Storage)`
    `SSA -> Wait-Graph Monitor -> Deadlock Detection`
* **APIs / Interfaces:**
    * `POST /ssa/lock`: Request a key-specific lock with a timeout.
    * `POST /ssa/commit`: Atomically update a key and release its lock.
* **Data Storage/State:**
    In-memory wait-graph for active locks; SQLite for versioned KV storage.

## 5. Alternatives Considered
* **Implicit Locking in SQLite:** Rejected because it doesn't provide the semantic visibility needed for agentic "Reasoning Loops" detection.
* **Global Swarm Lock:** Rejected due to extreme performance degradation in parallel meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Locks are tied to the "Mission Root" token; agents cannot lock keys outside their authorized scope.
* **Observability:** SSA logs "Wait-Time" and "Conflict Frequency" to the security dashboard.

## 7. Evolutionary Changelog
* **2026-03-21:** Initial Document Creation.
