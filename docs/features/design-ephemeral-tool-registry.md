# Design Doc: Ephemeral Tool Registry
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Modern agent swarms frequently generate task-specific tools (e.g., custom regex parsers, data formatters) that only need to exist for a short duration. Registering these through the traditional file-based configuration reload cycle is too slow (500ms-2s) and pollutes the permanent tool registry. The Ephemeral Tool Registry provides a high-speed, in-memory store for these transient tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Enable tool registration/unregistration in <10ms.
    * Automatically scope ephemeral tools to a specific session or subagent.
    * Support "Self-Destruct" timers for tools.
* **Non-Goals:**
    * Persisting ephemeral tools across server restarts.
    * Replacing the primary `config.yaml` for stable system tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Generate and use a custom data-cleaning tool for a 5-minute task.
* **The Happy Path (Tasks):**
    1. Orchestrator calls `register_ephemeral_tool` with the tool schema and execution logic (e.g., a Python snippet or a WASM module).
    2. MCP Any validates the tool and adds it to the session-scoped registry.
    3. Agent performs its task using the new tool.
    4. Task completes; MCP Any automatically prunes the tool based on the `ttl` (Time-To-Live) or session closure.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> register_ephemeral_tool -> In-Memory Registry -> Tool Execution (WASM/Sandbox) -> Cleanup`
* **APIs / Interfaces:**
    * `mcp_register_ephemeral_tool(name, schema, runtime, code, ttl_seconds)`
    * `mcp_unregister_ephemeral_tool(name)`
* **Data Storage/State:**
    * In-memory map (Concurrent Map) indexed by `session_id:tool_name`.

## 5. Alternatives Considered
* **Dynamic Config Generation**: Too much disk I/O and overhead from reloading the entire server state.
* **Global Ephemeral Registry**: Risk of name collisions between different swarms. Session-scoping is mandatory.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Ephemeral tools run in a restricted sandbox (e.g., WASM or isolated Python process) to prevent host compromise.
* **Observability:** Ephemeral tools are marked with an `is_ephemeral` flag in the UI and traces.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
