# Design Doc: Context Interop Bridge
**Status:** Draft
**Created:** 2026-03-14

## 1. Context and Scope
With the proliferation of different agent frameworks (OpenClaw, Claude Code, Gemini CLI), a major challenge has emerged: **Context Silos**. Each framework manages its own context, memory, and state, making it impossible for a Claude Code agent to seamlessly hand off a task to an OpenClaw specialist without losing critical progress. The `Context Interop Bridge` aims to standardize context exchange via a pluggable, MCP-compliant API.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified API for reading and writing agent context.
    * Support "Intent-Preserving" context compression to mitigate "Context Ghosting."
    * Implement lifecycle hooks for frameworks to sync their internal state with MCP Any.
    * Support multiple backend providers (OpenClaw ContextEngine, SQLite, Redis).
* **Non-Goals:**
    * Replacing the internal memory systems of agent frameworks.
    * Defining a single, mandatory context format (it should be an extensible schema).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator.
* **Primary Goal:** Sync state between a research agent (Claude Code) and a coding agent (OpenClaw).
* **The Happy Path (Tasks):**
    1. Research agent completes a task and writes its "Summary of Findings" to the Context Interop Bridge.
    2. The Bridge applies an "Intent-Preserving" compression hook to ensure the core goal is highlighted.
    3. The Orchestrator triggers the coding agent.
    4. The coding agent calls the `context/read` tool from MCP Any.
    5. The Bridge retrieves the compressed findings and injects them into the coding agent's prompt.
    6. The coding agent continues the work with full historical context.

## 4. Design & Architecture
* **System Flow:**
    `Agent A` -> `Context Write API` -> `Lifecycle Hooks (Compress/Filter)` -> `State Store` -> `Context Read API` -> `Agent B`
* **APIs / Interfaces:**
    * `context/get_session(session_id)`: Retrieve full or partial session state.
    * `context/update_session(session_id, delta)`: Append new context or state.
    * `context/register_hook(type, endpoint)`: Register a remote lifecycle hook (e.g., for OpenClaw ContextEngine).
* **Data Storage/State:**
    * Pluggable backends: Defaults to the existing "Blackboard" (SQLite).

## 5. Alternatives Considered
* **Centralized Vector DB**: Too heavy for simple state handoffs; better for long-term memory.
* **Manual Prompt Injection**: Error-prone and doesn't scale with high-frequency swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Context is "Intent-Bound." Agents can only read context that matches their authorized scope.
* **Observability**: A "Context Graph" UI will visualize how state flows between different agents.

## 7. Evolutionary Changelog
* **2026-03-14:** Initial Document Creation.
