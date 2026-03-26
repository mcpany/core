# Design Doc: LFMC (Lock-Free Mesh Coordination)

**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope

As swarms scale horizontally, global coordination locks become a bottleneck.
LFMC introduces a lock-free, CRDT-based coordination bus that allows teammates
to update their state fragments without global synchronization.

## 2. Goals & Non-Goals

* **Goals:**
  * Enable high-concurrency teammate coordination without global state locks.
  * Synchronize teammate task lists using eventual consistency (CRDTs).
  * Provide deterministic conflict resolution for overlapping task claims.
* **Non-Goals:**
  * Enforcing global transaction ordering (prioritizes availability over
    strict linearizability).

## 3. Critical User Journey (CUJ)

* **User Persona:** Specialized Subagent Teammate
* **Primary Goal:** Claim a sub-task and update status without waiting for a
  global coordination heartbeat.
* **The Happy Path (Tasks):**
  1. Subagent A identifies an available sub-task in the shared mesh.
  2. Subagent A issues a "Claim" fragment via the CRDT bus.
  3. Subagent B simultaneously attempts to claim the same task.
  4. Both claims are propagated; LFMC reconciles them deterministically using
     hardware timestamps.
  5. One agent succeeds, the other automatically pivots to the next available
     task.

## 4. Design & Architecture

* **System Flow:**
  * Subagent Claim -> CRDT State Update -> Mesh Propagation -> Deterministic
    Reconciliation -> Local Cache Sync.
* **APIs / Interfaces:**
  * `POST /v1/mesh/fragment/claim`: Claim a specific task fragment.
  * `GET /v1/mesh/state`: Retrieve the reconciled mesh state.
* **Data Storage/State:**
  * State fragments are stored in a distributed, lock-free blackboard.

## 5. Alternatives Considered

* **Redis-based Global Locking:** Rejected due to the 50ms+ latency tax and
  single-point-of-failure risks in deep, distributed meshes.

## 6. Cross-Cutting Concerns

* **Security (Zero Trust):** All CRDT fragments must be HAIL-attested to
  prevent state-injection by rogue agents.
* **Observability:** Real-time state-graph visualization in the
  Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog

* **2026-06-19:** Initial Document Creation.
