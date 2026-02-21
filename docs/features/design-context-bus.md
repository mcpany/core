# Design Doc: Inter-agent Context Bus
**Status:** Draft
**Created:** 2026-02-21

## 1. Context and Scope
As AI agent swarms (like those powered by AutoGen or Claude Code) become more common, a recurring problem is the fragmentation of context. When a parent agent spawns a subagent, the subagent often lacks the session-specific credentials, environment variables, or behavioral constraints of the parent.

MCP Any, acting as a gateway, is uniquely positioned to solve this by providing a unified "Context Bus" where state can be stored and inherited by any agent participating in the session.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized way to propagate auth headers from parent to subagents.
    * Allow agents to share "session memory" (e.g., current project root, active issue ID).
    * Enforce consistent security policies across an entire swarm.
* **Non-Goals:**
    * Implementing a full-blown long-term memory database (Vector DB).
    * Managing inter-agent message passing (A2A) beyond context sharing.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Share secure context between 3 agents without exposing local env vars to each agent individually.
* **The Happy Path (Tasks):**
    1. The Orchestrator initializes a session with MCP Any, providing a `session_id`.
    2. The Orchestrator sets "Session Context" (e.g., `GITHUB_TOKEN`, `PROJECT_ROOT`) via a new MCP Any management tool.
    3. The Orchestrator spawns 3 subagents, all referencing the same `session_id`.
    4. When Subagent A calls a tool (e.g., `list_repos`), MCP Any automatically injects the `GITHUB_TOKEN` from the session context.
    5. Subagent B inherits the `PROJECT_ROOT` and restricts its file operations accordingly.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Orchestrator -->|SetContext| Bus[Context Bus Service]
        Bus -->|Store| DB[(Session State)]
        SubagentA -->|CallTool| Gateway[MCP Any Gateway]
        Gateway -->|LookupContext| Bus
        Bus -->|Auth/Env| Gateway
        Gateway -->|ProxiedRequest| Upstream[Upstream API]
    ```
* **APIs / Interfaces:**
    * `context/set`: MCP Tool to define session-wide context.
    * `context/get`: Retrieve current session context (restricted).
    * `session_id` header in MCP JSON-RPC requests.
* **Data Storage/State:**
    * In-memory cache for active sessions, backed by SQLite for persistence and audit logs.

## 5. Alternatives Considered
* **Client-side Context Passing:** Rejected because it forces the AI models to handle sensitive tokens directly, increasing risk of exposure via prompt injection or log leakage.
* **Environment Variable Syncing:** Rejected as it's not dynamic enough for multi-tenant or multi-session use cases on the same host.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Context items can be marked as `sensitive`. Sensitive items are never returned to the agent; they are only injected by the gateway into upstream requests.
* **Observability:** Session-based logging allows for "Swarm Tracing," showing how context propagated from parent to child.

## 7. Evolutionary Changelog
* **2026-02-21:** Initial Document Creation.
