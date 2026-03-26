# Design Doc: Swarm Observability Hub (SOB)
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
With the rise of horizontal "Agent Teams" (e.g., Claude Code), agents are increasingly working in parallel across multiple task panels. Current observability tools are optimized for linear sessions, making it difficult to visualize the collective progress and inter-agent coordination of a swarm. SOB provides a high-speed, sharded telemetry service that aggregates and distributes real-time state updates for multi-panel visualization.

## 2. Goals & Non-Goals
* **Goals:**
    * Aggregate real-time telemetry from parallel teammates into a unified "Swarm View."
    * Provide sharded telemetry export to support multi-panel (e.g., tmux-like) visualization.
    * Track "Shared Task List" progression across disparate framework UIs.
    * Minimize telemetry latency to ensure real-time panel synchronization.
* **Non-Goals:**
    * Replacing framework-specific internal logs (e.g., Claude Code's debug logs).
    * Providing long-term archival of all reasoning traces (this is handled by the Telemetry Sink).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Monitor 3 parallel teammates refactoring a project and see their coordinated progress in real-time.
* **The Happy Path (Tasks):**
    1. The orchestrator spawns an agent team via Claude Code.
    2. MCP Any initializes a Swarm Session in the SOB.
    3. As teammates claim and complete tasks, they push lightweight state fragments to the SOB.
    4. The SOB aggregates these updates and streams them to the Swarm Observability UI.
    5. The UI displays three parallel panels, each showing a teammate's current task, reasoning status, and contribution to the shared task list.
    6. The orchestrator identifies a coordination stall and intervenes via the mission-root panel.

## 4. Design & Architecture
* **System Flow:**
    * Parallel Agents -> SOB Gateway -> Sharded Telemetry Buffer -> Real-time UI Stream.
* **APIs / Interfaces:**
    * `PublishSwarmState(teammateId, stateFragment)`: Pushes lightweight JSON updates.
    * `SubscribeSwarmStream(missionId)`: WebSocket/SSE endpoint for UI synchronization.
* **Data Storage/State:** High-speed, sharded in-memory buffer (Redis/Memcached style) with periodic snapshots to SQLite.

## 5. Alternatives Considered
* **Centralized Database Polling**: Rejected due to high latency and database contention in high-density swarms.
* **Direct P2P Telemetry**: Rejected because it increases the cognitive load and network overhead for individual specialist agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Telemetry fragments are origin-locked and mission-bound.
* **Observability:** SOB is self-observing, tracking its own message throughput and buffer health.

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
