# Design Doc: Intent-Scoped Snapshot Controller

**Status:** Draft
**Created:** 2026-05-03

## 1. Context and Scope
As agent swarms become more parallel and speculative, global environment rollbacks (via PLSS) are becoming too disruptive. A failure in one specialized intent branch should not force the entire swarm to lose progress. The Intent-Scoped Snapshot Controller extends the PLSS bridge to provide "targeted" rollbacks, ensuring environment resilience with minimal blast radius.

## 2. Goals & Non-Goals
* **Goals:**
    * Perform environment rollbacks restricted to a specific `mission_id` or `intent_id` branch.
    * Integrate with the UACO v2.2 Intent Barrier Middleware to detect conflict-free snapshot boundaries.
    * Support "Shadow-FS" path merging for speculative results.
* **Non-Goals:**
    * Managing OS-level volume snapshots (delegated to the underlying PLSS bridge).
    * Resolving semantic merge conflicts in shared files (delegated to the agent).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Concurrency Swarm Architect
* **Primary Goal:** Recover from a failed speculative file edit without interrupting three other parallel sub-intents.
* **The Happy Path (Tasks):**
    1. Sub-Intent B initiates a high-risk file transformation.
    2. The Snapshot Controller creates an "Intent-Bound" snapshot of the affected filesystem region.
    3. Sub-Intent B fails a security quorum or attestation check.
    4. The Controller triggers an atomic rollback restricted to the paths modified by Sub-Intent B.
    5. Parallel Sub-Intents A and C continue their work uninterrupted.

## 4. Design & Architecture
* **System Flow:**
    `[Intent Event] -> [Path-Intent Tracker] -> [PLSS Proxy] -> [Shadow-FS Rollback]`
* **APIs / Interfaces:**
    * `CreateIntentSnapshot(intent_id string, paths []string) (SnapshotID, error)`
    * `RollbackIntent(intent_id string) error`
* **Data Storage/State:**
    Maintains a mapping of `intent_id` to modified `Inode` sets and `Shadow-FS` overlays.

## 5. Alternatives Considered
* **Global Rollback:** Simple but disruptive. Rejected for high-concurrency swarms.
* **File-by-File Backup:** Inefficient for large directories. Rejected in favor of Inode-based tracking.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Snapshot boundaries must be cryptographically bound to the signed intent to prevent "Snapshot Splicing."
* **Observability:** Visualize rollback events in the "Swarm Rollback Dashboard."

## 7. Evolutionary Changelog
* **2026-05-03:** Initial Document Creation.
