# Design Doc: A2A Task Lifecycle Manager

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As the A2A (Agent-to-Agent) protocol matures, agents are moving away from synchronous request-response cycles towards long-running, asynchronous tasks. A parent agent might delegate a task that takes minutes or hours to complete. MCP Any must act as the stateful orchestrator that tracks these tasks, handles intermediate "input required" states, and ensures results are delivered even if the parent agent disconnects.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a state machine for A2A Task objects (submitted, working, input-required, completed, failed).
    *   Provide a "Task Buffer" where agents can poll for status or receive webhook updates.
    *   Expose task management tools (e.g., `get_task_status`, `provide_task_input`) to MCP-native agents.
    *   Ensure task state persistence across server restarts.
*   **Non-Goals:**
    *   Defining the task logic itself (this is the responsibility of the executing agent).
    *   Implementing a full-blown workflow engine (like Temporal). This is a lightweight task bridge.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Distributed Agent Swarm Orchestrator.
*   **Primary Goal:** Manage a multi-step task where a "Researcher" agent needs human input halfway through.
*   **The Happy Path (Tasks):**
    1.  Agent A submits a task to Agent B via MCP Any: `delegate_task(target="AgentB", payload={...})`.
    2.  MCP Any creates a Task ID `task_123` and sets status to `submitted`.
    3.  Agent B picks up the task and updates status to `working`.
    4.  Agent B encounters a blocker and updates status to `input-required`.
    5.  MCP Any notifies Agent A (or a human via UI).
    6.  Input is provided via `provide_task_input(task_id="task_123", input={...})`.
    7.  Agent B completes the task; MCP Any stores the final result and marks it `completed`.

## 4. Design & Architecture
*   **System Flow:**
    - **Task Registry**: A central middleware component that intercepts A2A-bound tool calls and initializes task state in the `Shared KV Store`.
    - **Event Bus**: An internal pub/sub system that triggers notifications when task states transition.
    - **A2A Adapter**: The outbound layer that communicates task updates to external A2A agents.
*   **APIs / Interfaces:**
    - `mcp_a2a_create_task(agent_id, payload) -> task_id`
    - `mcp_a2a_get_task(task_id) -> TaskStatus`
    - `mcp_a2a_update_task(task_id, status, metadata)`
*   **Data Storage/State:** Task states are persisted in the embedded SQLite `Blackboard` with a TTL (Time-To-Live) to prevent storage bloat.

## 5. Alternatives Considered
*   **Stateless Proxying**: Just pass messages back and forth. *Rejected* because it fails when agents are intermittent or tasks are long-running.
*   **Using an External DB**: Requiring Redis/Postgres. *Rejected* to keep MCP Any's "Local-First" and "Zero-Dependency" philosophy.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Only the task "Owner" (creator) or the "Assignee" (executor) can update task state. Access is verified via capability tokens.
*   **Observability:** The UI "Agent Chain Tracer" will use Task IDs to group related tool calls and messages into a single logical execution flow.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
