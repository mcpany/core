<!-- markdownlint-disable -->
# Design Doc: LFMC (Lock-Free Mesh Coordination)
**Status:** Draft
**Created:** 2026-06-19

## 1. Context and Scope
LFMC introduces a lock-free, CRDT-based coordination bus that allows teammates to update their state fragments without global synchronization.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable lock-free teammate coordination in horizontal swarms.
    * Synchronize teammate task lists without global coordination locks.
    * Use CRDT-based state fragments to prevent write-collisions.
* **Non-Goals:**
    * Eliminating all coordination (only "Lock-Free" at the shard level).
    * Enforcing global ordering (only eventual consistency for non-conflicting tasks).

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialized Subagent Teammate
* **Primary Goal:** Update its task status and claim a sub-task without waiting for a global coordination lock.
* **The Happy Path (Tasks):**
    1. Subagent A identifies an available sub-task in the LFMC mesh.
    2. Subagent A claims the task using a CRDT-based "Claim" fragment.
    3. Subagent B simultaneously attempts to claim a different task.
    4. Both claims are propagated across the mesh without a global lock.
    5. The claims are reconciled using deterministic LWW (Last-Write-Wins) logic.
    6. Both agents begin their respective tasks.

## 4. Design & Architecture
* **System Flow:**
    * Subagent Claim -> CRDT State Update -> Mesh Propagation -> Deterministic Reconciliation -> Local Cache Sync.
* **APIs / Interfaces:**
    * `POST /v1/mesh/fragment/claim`: Claim a task fragment.
    * `GET /v1/mesh/state`: Retrieve the reconciled mesh state.
* **Data Storage/State:**
    * Replicated state fragments are stored in a distributed, lock-free blackboard.

## 5. Alternatives Considered
* **Redis-based Global Locking:** Rejected due to the high latency and single-point-of-failure risk in deep, distributed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mesh fragments are hardware-attested and session-bound.
* **Observability:** Integrated with the Mesh-Resident Lineage Tracker.

## 7. Evolutionary Changelog
* **2026-06-19:** Initial Document Creation.
