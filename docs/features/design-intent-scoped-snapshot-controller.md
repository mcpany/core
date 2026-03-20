# Design Doc: Intent-Scoped Snapshot Controller
**Status:** Draft
**Created:** 2026-05-03

## 1. Context and Scope
As agent swarms grow in complexity, global environment rollbacks become increasingly disruptive. A single subagent failure currently triggers a full project-level revert, wiping out progress made by other healthy, concurrent intent branches. MCP Any needs a mechanism to perform targeted rollbacks that only affect the specific files and state associated with a failed intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Snapshot Controller" that tracks file changes per intent branch.
    * Enable atomic rollbacks restricted to an intent-specific "Change Set."
    * Integrate with the PLSS (Project-Local Snapshot Sync) bridge for efficient storage.
* **Non-Goals:**
    * This system WILL NOT provide a full version control system (like Git).
    * It WILL NOT manage conflicts between overlapping intent branches (handled by the Parallel Intent Branch Manager).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars, and recover from a single agent's failure without affecting others.
* **The Happy Path (Tasks):**
    1. Orchestrator starts three parallel intent branches (A, B, C).
    2. Intent Branch B fails a security quorum or tool execution.
    3. The Snapshot Controller identifies the files modified by Branch B.
    4. The system performs a targeted rollback of Branch B's changes.
    5. Branches A and C continue execution without interruption or state loss.

## 4. Design & Architecture
* **System Flow:**
    The Snapshot Controller hooks into the Shadow-FS and PLSS. Every file write is tagged with an `intent_id`.
* **APIs / Interfaces:**
    * `CreateSnapshot(intent_id)`: Initializes a tracking session for an intent.
    * `RollbackIntent(intent_id)`: Reverts all changes tagged with the given ID.
    * `CommitIntent(intent_id)`: Merges the changes into the host/main project state.
* **Data Storage/State:**
    Uses an internal SQLite "Intent Ledger" to track Hardware Inode mappings and file hashes per intent.

## 5. Alternatives Considered
* **Global LVM/ZFS Snapshots:** Rejected due to lack of granularity; cannot rollback specific files while keeping others.
* **Git Branching:** Rejected as too slow and heavy for high-frequency subagent reasoning loops.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The controller must verify the `intent_id` against the signed mission intent to prevent "Cross-Intent Poisoning."
* **Observability:** Logs of all targeted rollbacks will be emitted to the mission audit trail.

## 7. Evolutionary Changelog
* **2026-05-03:** Initial Document Creation.
