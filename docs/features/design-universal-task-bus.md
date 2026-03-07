# Design Doc: Universal Task Bus (UTB) Protocol
**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
As multi-agent swarms (e.g., OpenClaw, Claude Code, AutoGen) become the primary way to build complex software, there is a growing need for a standardized "Universal Task Bus." Currently, each agent framework maintains its own internal task state and progress trackers. This creates a "State Silo" problem where different agents cannot easily hand off tasks or share a single, verifiable source of truth for task progress.

MCP Any, as the universal adapter and gateway, is uniquely positioned to provide this UTB layer. By standardizing the interface for task state, progress, and handoffs, MCP Any can allow disparate agent frameworks to coordinate more effectively, enabling a "Unified Agent Board" (e.g., Vibe-Kanban) to track the entire lifecycle of a complex development task.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized, MCP-compatible interface for creating, updating, and querying task status.
    * Enable "Task Handoff" messages between different agent frameworks (e.g., Claude Code hands off a bug fix task to OpenClaw for testing).
    * Provide a verifiable audit log of task progress and agent-to-agent interactions.
    * Support "State Persistence" for long-running, asynchronous tasks across agent restarts.
* **Non-Goals:**
    * Replacing existing agent-specific task managers (e.g., OpenClaw's internal state).
    * Providing a full-blown project management tool (e.g., Jira, Trello).
    * Dictating how an agent should execute its internal tasks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., a developer using both Claude Code and OpenClaw in a single project).
* **Primary Goal:** Coordinate a complex bug fix that requires initial research by Claude Code, followed by a formal fix and testing by OpenClaw.
* **The Happy Path (Tasks):**
    1.  The user initiates a bug fix task via Claude Code.
    2.  Claude Code creates a new "Task" object in the MCP Any UTB.
    3.  Claude Code completes its research and updates the UTB task status to `Research Complete`.
    4.  Claude Code initiates a "Handoff" to OpenClaw via the UTB, including a pointer to the research findings.
    5.  OpenClaw receives the handoff notification from the UTB and picks up the task.
    6.  OpenClaw implements the fix, runs tests, and updates the UTB task status to `Fix Verified`.
    7.  The user views a "Unified Agent Board" (powered by the UTB) that shows the entire history and current status of the bug fix.

## 4. Design & Architecture
* **System Flow:**
    * **Task Storage**: A persistent, SQLite-backed store within MCP Any to hold task metadata, status, and history.
    * **MCP Tool Interface**: MCP Any exposes tools for `utb_create_task`, `utb_update_task`, `utb_get_task_status`, and `utb_initiate_handoff`.
    * **Event Stream**: A WebSocket-based event stream for agents to subscribe to task updates and handoff notifications.
* **APIs / Interfaces:**
    * `utb_create_task(title: string, description: string, initial_agent_id: string) -> task_id: string`
    * `utb_update_task(task_id: string, status: string, progress_notes: string)`
    * `utb_get_task_status(task_id: string) -> task_object`
    * `utb_initiate_handoff(task_id: string, target_agent_id: string, handoff_context: JSON)`
* **Data Storage/State:**
    * Task state is stored in the same internal SQLite database used by the "Shared KV Store" and "A2A Stateful Residency" features.

## 5. Alternatives Considered
* **Agent-Specific APIs**: Rejected because it requires each agent framework to implement multiple adapters for every other framework.
* **Third-Party Task Managers (e.g., Jira API)**: Rejected because it introduces excessive latency and complexity for low-level agent-to-agent coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):**
    * Every UTB request must be signed by the originating agent's cryptographic key.
    * "Intent-Aware" permissions ensure that an agent can only update tasks it is authorized to work on.
* **Observability:**
    * All UTB actions are logged to the MCP Any audit log.
    * A dedicated UTB Dashboard in the MCP Any UI provides a real-time view of all active tasks and handoffs.

## 7. Evolutionary Changelog
* **2026-03-07:** Initial Document Creation.
