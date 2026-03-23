# Design Doc: Structured Task Event Bridge
**Status:** Draft
**Created:** 2026-05-09

## 1. Context and Scope
As AI agent swarms evolve from simple tool-calling chains to complex, long-running mission orchestrations, the need for a standardized "Completion Signal" has become critical. Today, disparate frameworks (OpenClaw, AutoGen, CrewAI) use incompatible methods to signal task success, failure, or handoff. MCP Any, as the Universal Agent Bus, is uniquely positioned to act as the authoritative event bridge that standardizes these signals into a unified `task_completion` event stream.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Standardize subagent runtime events into a typed `task_completion` format compatible with OpenClaw 2026.3.1.
    *   Provide an immutable, audit-ready event store for mission outcomes.
    *   Enable cross-framework event routing (e.g., an OpenClaw subagent triggering an AutoGen supervisor).
*   **Non-Goals:**
    *   Implementing a full task orchestration engine (MCP Any remains a bridge/gateway).
    *   Managing agent internal state (handled by the Blackboard).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Swarm Architect
*   **Primary Goal:** Verify mission completion and collect outcome data from a heterogeneous swarm of 5+ specialized agents.
*   **The Happy Path (Tasks):**
    1.  Architect configures the Structured Task Event Bridge in `mcpany`.
    2.  A specialized "Researcher" agent (OpenClaw) completes its task and emits a `task_completion` event to MCP Any.
    3.  MCP Any validates the event against the mission root and persists it to the immutable outcome store.
    4.  A "Writer" agent (AutoGen) receives the completion signal via the MCP Any A2A bridge and begins the next phase.
    5.  Architect views the unified mission timeline in the MCP Any UI, seeing the verified outcome from both agents.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        AgentA[Agent A - OpenClaw] -->|Emit Event| Bridge[Structured Task Event Bridge]
        Bridge -->|Validate & Persist| Store[(Immutable Outcome Store)]
        Bridge -->|Route| AgentB[Agent B - AutoGen]
        Store -->|Read| UI[MCP Any Dashboard]
    ```
*   **APIs / Interfaces:**
    *   `POST /events/task_completion`: Endpoint for agents to signal completion.
    *   `GET /mission/:id/timeline`: Retrieve the standardized event stream for a mission.
*   **Data Storage/State:**
    *   Events are stored in a dedicated, append-only table in the internal SQLite database.
    *   Each event is cryptographically linked to the parent `MissionRoot` and the agent’s `IdentityToken`.

## 5. Alternatives Considered
*   **Framework-Specific Adapters**: Rejected due to high maintenance overhead and lack of cross-framework standardization.
*   **Generic Webhook Bridge**: Rejected because it lacks the semantic validation and mission-linking required for secure agent swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All completion events must be signed by the agent’s identity token. MCP Any verifies that the agent had the authorized "Intent-Scope" to complete the specified task.
*   **Observability:** Integrated with the "Reasoning Effort Visualizer" to correlate completion latency with reported cognitive effort.

## 7. Evolutionary Changelog
*   **2026-05-09:** Initial Document Creation.
