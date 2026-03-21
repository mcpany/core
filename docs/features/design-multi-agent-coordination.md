# Design Doc: Multi-Agent Coordination System
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
As AI agent ecosystems evolve from single-agent monoliths to multi-agent swarms (e.g., OpenClaw, CrewAI), there is a critical need for a stable infrastructure layer that manages state and tool access across these agents. Currently, handoffs between agents often result in context loss or redundant tool calls. MCP Any will solve this by acting as a "Session-Aware Coordination Hub."

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified session ID that persists across multiple agents.
    * Track tool execution state to prevent redundant calls.
    * Enable secure "handoff" of context between agents using standardized headers.
* **Non-Goals:**
    * Replacing existing agent frameworks (CrewAI, AutoGen).
    * Orchestrating the actual logic of which agent to call next (this remains with the orchestrator).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator
* **Primary Goal:** Coordinate a research agent and a writing agent using shared tools without losing state.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a session with MCP Any.
    2. Research agent calls `search_tool` via MCP Any; results are cached in the session state.
    3. Research agent signals "handoff" to Writing agent.
    4. Writing agent resumes the session and accesses `search_tool` results directly from MCP Any's shared state without re-running the tool.

## 4. Design & Architecture
* **System Flow:**
    `Orchestrator -> Session Manager (MCP Any) -> [Agent A | Agent B] -> Tool Registry -> MCP Servers`
* **APIs / Interfaces:**
    * `POST /session/init`: Creates a new multi-agent session.
    * `POST /session/{id}/handoff`: Securely transfers context headers to a new agent ID.
    * `GET /session/{id}/state`: Retrieves current shared state/tool results.
* **Data Storage/State:**
    * State is managed in an embedded SQLite "Blackboard" per session, ensuring persistence and isolation.

## 5. Alternatives Considered
* **Client-Side State Management**: Rejected because it places too much burden on the agent framework and leads to inconsistent state across different frameworks.
* **Global Shared Cache**: Rejected because it lacks the necessary isolation and security boundaries required for multi-tenant or multi-task swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Sessions are bound to specific capability tokens. Agent B can only access Agent A's results if specifically authorized during the handoff.
* **Observability**: Session logs will provide a "Timeline View" of which agent called which tool and when the handoff occurred.

## 7. Evolutionary Changelog
* **2026-02-24**: Initial Document Creation.
