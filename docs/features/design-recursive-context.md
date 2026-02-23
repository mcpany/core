# Design Doc: Recursive Context Protocol (RCP)
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As agentic workflows evolve from single-agent tasks to multi-agent swarms (e.g., a parent agent orchestrating specialized subagents), a significant challenge arises: **Context Fragmentation**. Subagents often lack the high-level intent, conversational history, or security context of the parent, leading to redundant queries, hallucinations, or security bypasses. MCP Any is uniquely positioned to bridge this gap by standardizing how context is inherited through the Model Context Protocol.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize context inheritance via custom MCP headers (`X-MCP-Parent-Context-ID`).
    * Implement "Intent-Scoped" state sharing between parent and child agents.
    * Enable automatic context injection for subagents without manual configuration.
* **Non-Goals:**
    * Building a full agent framework (MCP Any remains an adapter/gateway).
    * Storing massive amounts of raw conversation history (focus on structured state and intent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator (using CrewAI/OpenClaw).
* **Primary Goal:** Share secure context between a "Manager" agent and 3 "Worker" agents without exposing full environment variables or sensitive keys.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a session with MCP Any.
    2. Parent agent spawns a subagent and passes a "Context Token".
    3. Subagent calls a tool via MCP Any using the Context Token.
    4. MCP Any automatically injects relevant parent state (e.g., current project directory, shared variables) into the tool execution environment.
    5. Subagent completes the task with full awareness of the manager's intent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Parent Agent->>MCP Any: Create Session (Initial Context)
        MCP Any-->>Parent Agent: Session ID / Token
        Parent Agent->>Subagent: Delegate Task + Token
        Subagent->>MCP Any: Tool Call (X-MCP-Session-Token: Token)
        MCP Any->>Context Store: Fetch Parent State
        Context Store-->>MCP Any: State (Merged Context)
        MCP Any->>Tool Upstream: Execute Tool with Injected Context
    ```
* **APIs / Interfaces:**
    * `POST /context/session`: Initialize a new context session.
    * `GET /context/session/:id`: Retrieve state (internal use).
    * Custom Headers: `X-MCP-Session-ID`, `X-MCP-Context-Scope`.
* **Data Storage/State:**
    * Shared state managed in the embedded SQLite "Blackboard" (`Shared KV Store`).

## 5. Alternatives Considered
* **Manual Context Passing:** Rejected because it's error-prone and leads to "Context Bloat" in the LLM window.
* **Global Environment Variables:** Rejected because it violates the principle of least privilege and lacks session isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are time-bound and scope-restricted. A subagent cannot escalate its context access beyond what the parent explicitly granted.
* **Observability:** Trace IDs will link parent and child tool calls in the Tool Execution Timeline.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation. Standardizing Recursive Context Protocol for multi-agent swarm orchestration.
