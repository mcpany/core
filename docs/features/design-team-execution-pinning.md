# Design Doc: Team Execution Pinning Middleware
**Status:** Draft
**Created:** 2026-05-15

## 1. Context and Scope
With the rise of parallel multi-agent swarms (e.g., Claude Code "Agent Teams"), multiple agents now operate simultaneously within the same project environment. Without strict coordination, parallel teammates often inadvertently overwrite each other's work or operate on stale file states, leading to "State Divergence." MCP Any needs a middleware layer for the Filesystem Adapter that can enforce sub-tree locks and ensure that specific directories are pinned to specific subagents during their execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Pinning Proxy" for the Filesystem Adapter.
    * Enforce sub-tree isolation for parallel teammate subagents.
    * Prevent race conditions and state divergence in shared project filesystems.
    * Provide a standardized lock-acquisition protocol for agent frameworks.
* **Non-Goals:**
    * Providing a general-purpose distributed file lock (e.g., NFS locks).
    * Managing file versioning or git-like branching (handled by PLSS).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Coordinate 3 parallel agents (Frontend, Backend, Tests) working on the same repository without file write conflicts.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns 3 subagents and assigns them to the "Agent Team" mission.
    2. Frontend Agent requests a pin for `ui/src/`.
    3. MCP Any validates the request against the mission root and grants the exclusive sub-tree lock.
    4. Backend Agent attempts to write to `ui/src/` but is blocked by the Pinning Middleware.
    5. Frontend Agent completes its task and releases the pin.
    6. The lock is automatically purged upon subagent session termination.

## 4. Design & Architecture
* **System Flow:**
    `[Subagent] -> [Filesystem Adapter] -> [Team Execution Pinning Middleware] -> [Host Filesystem]`
* **APIs / Interfaces:**
    * `PinningService.AcquirePin(path, subagent_id, mission_id)`: Acquires an exclusive lock on a sub-tree.
    * `PinningService.ReleasePin(path, subagent_id)`: Explicitly releases a lock.
* **Data Storage/State:**
    * In-memory Lock Table (Radix Tree) mapping paths to `subagent_id`.
    * Persistence for long-running missions in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **Agent-Side Locking:** Rejected as it relies on agent compliance and cannot be enforced across different frameworks.
* **Full OS-Level Namespacing (Containers):** Rejected due to the high overhead and complexity of sharing project-local dependencies across containers.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Pinning requests must be cryptographically signed by the parent agent to prevent unauthorized lock-squatting.
* **Observability:** Active pins and lock contention events are visualized in the "Parallel Intent Visualizer."

## 7. Evolutionary Changelog
* **2026-05-15:** Initial Document Creation.
