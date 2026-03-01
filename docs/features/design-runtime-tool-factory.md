# Design Doc: Runtime Tool Factory (Dynamic Synthesis)
**Status:** Draft
**Created:** 2026-03-01

## 1. Context and Scope
With the rise of agent frameworks like OpenClaw, agents are moving away from static toolsets. Modern swarms require the ability to "synthesize" specialized subagents with highly specific, temporary tools. MCP Any currently treats tool registration as a semi-static configuration process. This document proposes a "Runtime Tool Factory" that allows agents to define, scope, and register tools on-the-fly, bound to a specific session or intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Allow agents to dynamically register temporary MCP tools via a standardized factory tool.
    * Implement "Intent-Bound Scoping" where synthesized tools are only accessible to a specific subagent or session.
    * Provide a mechanism for "Ephemeral Tool Lifecycle" (tools that expire after a task or timeout).
    * Integrate with the Policy Engine to ensure synthesized tools do not bypass security perimeters.
* **Non-Goals:**
    * Writing the actual logic for the upstream services (the factory registers wrappers for existing or newly spawned ephemeral services).
    * Replacing static configuration for core infrastructure tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** OpenClaw Swarm Orchestrator
* **Primary Goal:** Create a specialized "Log Parser" subagent with a one-time tool to access a specific log file.
* **The Happy Path (Tasks):**
    1. Orchestrator identifies a need for log analysis on `auth-service-v2`.
    2. Orchestrator calls `mcpany_synthesize_tool` with a definition for `read_auth_logs` restricted to `/var/log/auth.log`.
    3. MCP Any validates the request against the Orchestrator's permissions.
    4. MCP Any registers the ephemeral tool and generates a unique `Session-Scope-ID`.
    5. Orchestrator spawns a Subagent, passing the `Session-Scope-ID`.
    6. Subagent calls `read_auth_logs`. MCP Any allows the call because the `Session-Scope-ID` matches.
    7. Once the task is complete, the Orchestrator calls `mcpany_teardown_tool` (or it expires).

## 4. Design & Architecture
* **System Flow:**
    - **Factory Interface**: A core MCP tool `mcpany_synthesize_tool` that accepts a tool schema and an upstream provider (e.g., a dynamic Docker container or a scoped CLI command).
    - **Session Registry**: A new in-memory (or Redis-backed) registry that maps `Session-Scope-IDs` to ephemeral tool definitions.
    - **Middleware Injection**: The factory automatically injects "Intent-Aware" middleware into the synthesized tool's execution pipeline.
* **APIs / Interfaces:**
    ```json
    {
      "method": "tools/call",
      "params": {
        "name": "mcpany_synthesize_tool",
        "arguments": {
          "name": "ephemeral_tool_name",
          "schema": { ...JSON Schema... },
          "upstream_config": { ...Adapter Config... },
          "ttl_seconds": 3600,
          "scope": "session_abc_123"
        }
      }
    }
    ```
* **Data Storage/State:** Ephemeral state managed by the `ServiceRegistry` with an "auto-expire" worker.

## 5. Alternatives Considered
* **Just-in-Time Upstreams**: Spawning a new MCP Any instance for every subagent. *Rejected* due to massive resource overhead and difficulty in centralizing audit logs.
* **Broad Static Permissions**: Giving agents access to all logs and letting them filter. *Rejected* as it violates the Principle of Least Privilege.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Synthesized tools must inherit the *intersection* of the parent agent's permissions and the requested scope. A subagent cannot be "synthesized" with more power than its parent.
* **Observability:** All synthesis events are recorded in the audit log with a reference to the parent `intent_id`.

## 7. Evolutionary Changelog
* **2026-03-01:** Initial Document Creation.
