# Design Doc: ContextEngine Lifecycle Adapter
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
As AI agents evolve into complex systems, managing conversation context (memory) becomes a primary bottleneck. Frameworks like OpenClaw have introduced pluggable "ContextEngines" to handle context lifecycle. MCP Any needs to provide a standardized adapter that allows any agent to offload these lifecycle hooks to a centralized, high-performance middleware. This ensures that context strategies (RAG, summarization, isolation) are reusable across different agent frameworks and environments.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized MCP-like interface for context lifecycle hooks (`bootstrap`, `ingest`, `assemble`, `compact`, `afterTurn`).
    * Enable cross-framework context sharing (e.g., OpenClaw agent and Claude Code sharing the same "Blackboard" memory).
    * Implement "Context Isolation" where subagents only see a restricted, "compacted" view of the parent's memory.
    * Support pluggable context backends (SQLite, Vector DB, Redis).
* **Non-Goals:**
    * Implementing specific LLM-based summarization logic (this is handled by the plugins).
    * Replacing the agent's internal short-term memory entirely.

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Developer.
* **Primary Goal:** Share a consistent, secure memory state between a specialized researcher agent and a writer agent.
* **The Happy Path (Tasks):**
    1. Developer configures a "Shared Context" resource in MCP Any.
    2. Researcher agent calls the `ingest` hook to save findings.
    3. MCP Any stores the findings in the Shared KV Store (Blackboard).
    4. Writer agent calls the `assemble` hook.
    5. MCP Any applies a `compact` plugin to summarize the researcher's findings and injects them into the writer's prompt context.

## 4. Design & Architecture
* **System Flow:**
    `Agent` <-> `ContextEngine Adapter (MCP Any)` <-> `Context Plugins` <-> `Shared KV Store (Blackboard)`
    1. **Hook Interception**: MCP Any exposes standard tools for each lifecycle hook.
    2. **Plugin Execution**: When a hook is called, MCP Any executes the registered plugins (e.g., a "Summarization Plugin" on `compact`).
    3. **State Persistence**: The final state is persisted in the Agent-Aware Blackboard with mandatory isolation.
* **APIs / Interfaces:**
    * `context_bootstrap(session_id, initial_state)`
    * `context_ingest(session_id, content, metadata)`
    * `context_assemble(session_id) -> prompt_snippet`
    * `context_compact(session_id, limit)`
* **Data Storage/State:**
    * Uses the existing **Shared KV Store** (design-agent-aware-blackboard.md) for persistence.

## 5. Alternatives Considered
* **Agent-Specific Plugins**: Rejected because it creates siloed memory that cannot be shared between frameworks (OpenClaw vs. Claude Code).
* **Direct Vector DB Access**: Too low-level; doesn't provide the lifecycle management needed for conversational AI.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All `ingest` and `assemble` calls are subject to the Policy Firewall. Context snippets are "Intent-Bound" to prevent cross-agent data leakage.
* **Observability**: The UI will provide a "Blackboard Isolation Inspector" to visualize memory state across different sessions.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
