# Design Doc: Transactional Blackboard (Atomic Swarm State)
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As AI agent swarms (e.g., OpenClaw, CrewAI) scale to dozens of concurrent sub-agents, they increasingly rely on a shared memory space (the "Blackboard") to coordinate. Currently, MCP Any provides a basic SQLite-backed KV store. However, without atomicity and locking, concurrent agents risk "lost updates" and state corruption when multiple agents attempt to update the same task status or context block simultaneously.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement row-level locking (mutex) for specific keys.
    * Provide atomic "Compare-and-Swap" (CAS) operations via the MCP tool interface.
    * Ensure "Agent-Bound" isolation persists during transactional operations.
* **Non-Goals:**
    * Implementing a full distributed SQL engine.
    * Support for multi-key cross-table transactions (focus is on per-key atomicity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (e.g., OpenClaw Lead Agent)
* **Primary Goal:** Safely increment a "completed_tasks" counter from 5 concurrent sub-agents without missing any increments.
* **The Happy Path (Tasks):**
    1. Agent A calls `blackboard_get(key="task_counter")`.
    2. Agent A calculates new value.
    3. Agent A calls `blackboard_cas(key="task_counter", old_value=5, new_value=6)`.
    4. If another agent updated it first, Agent A receives a `PRECONDITION_FAILED` error and retries.

## 4. Design & Architecture
* **System Flow:**
    * MCP Any intercepts `blackboard_*` tool calls.
    * Middleware checks if the requested key is "Locked" in the session state.
    * SQLite `UPDATE ... WHERE key=? AND value=?` is used for CAS operations.
* **APIs / Interfaces:**
    * `blackboard_cas(key, old_val, new_val)`
    * `blackboard_lock(key, ttl)`
    * `blackboard_unlock(key)`
* **Data Storage/State:**
    * SQLite table `blackboard` with columns `key`, `value`, `version`, `locked_by`, `lock_expires_at`.

## 5. Alternatives Considered
* **Redis:** Rejected to keep MCP Any dependency-free and "Local-First."
* **File-based locks:** Rejected due to performance overhead and lack of atomic CAS support in plain filesystems.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** CAS and Locks are still subject to "Agent-Bound" isolation. Agent B cannot lock or CAS a key owned by Agent A unless they share an "Intent Scope."
* **Observability:** Log "Lock Contention" metrics to help users identify bottlenecks in their swarm architecture.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
