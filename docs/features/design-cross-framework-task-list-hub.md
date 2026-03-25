# Design Doc: Cross-Framework Task List Hub
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
As AI agent ecosystems move from single-framework silos (e.g., only Claude Code or only OpenClaw) to heterogeneous swarms, the need for a unified coordination layer becomes critical. "Agent Teams" currently rely on framework-specific shared task lists. MCP Any needs to provide a universal, framework-agnostic Task List Hub that allows teammates from different frameworks to synchronize state, claim tasks, and report progress without global coordination locks.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified, sharded task list accessible via MCP, gRPC, and UACO.
    * Implement Conflict-Free Replicated Data Types (CRDTs) to ensure non-blocking task updates across parallel teammates.
    * Support hardware-attested task claiming to prevent hijacking by rogue subagents.
    * Enable cross-framework task delegation (e.g., a Claude-spawned lead delegating to an OpenClaw specialist).
* **Non-Goals:**
    * Replacing the agent's internal reasoning about task priority.
    * Providing a persistent long-term storage for completed tasks (handled by the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Spawning a multi-agent team comprising Claude Code lead and OpenClaw specialists to perform a complex refactor.
* **The Happy Path (Tasks):**
    1. The lead agent initializes a "Mission" and creates a shared task list in the Hub.
    2. The lead agent spawns two specialists from different frameworks.
    3. Specialists authenticate with the Hub using hardware-attested mission tokens.
    4. Specialists query the Hub for available tasks and "claim" them using a CRDT-based update.
    5. Specialists perform their tasks and update the Hub with status and result fragments.
    6. The lead agent monitors the Hub and merges results into the mission-root state.

## 4. Design & Architecture
* **System Flow:**
    `[Agent A (Claude)] <-> [Task List Hub (CRDTs)] <-> [Agent B (OpenClaw)]`
* **APIs / Interfaces:**
    * `createTaskList(missionId, initialTasks)`
    * `claimTask(missionId, taskId, agentToken)`
    * `updateTaskStatus(missionId, taskId, status, resultFragment)`
    * `subscribeToUpdates(missionId)` (WebSocket/gRPC stream)
* **Data Storage/State:**
    * In-memory CRDT state for active missions.
    * SQLite-backed persistence for mission checkpoints.

## 5. Alternatives Considered
* **Centralized Locking (Redis-style):** Rejected due to high latency and potential for deadlocks in high-density parallel swarms.
* **Framework-Specific Bridges:** Rejected as it requires N^2 adapters for every framework combination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory hardware-attestation for all task-claiming operations. Intent-bound isolation ensures agents only see tasks within their authorized scope.
* **Observability:** Real-time visualization of the shared task list and agent claiming patterns in the Multi-Agent Swarm Topology Monitor.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
