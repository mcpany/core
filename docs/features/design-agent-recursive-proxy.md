# Design Doc: Recursive Agent-as-a-Server Proxy
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
With the release of Claude Code's `mcp serve` capability, the ecosystem is shifting towards "Recursive Agency." This pattern involves a primary LLM agent delegating high-level, complex tasks (e.g., "Refactor this entire module") to another specialized agent instance, rather than calling primitive tools (e.g., `read_file`, `write_file`).

MCP Any needs to formalize this by acting as a **Recursive Proxy**. It must allow any MCP-compliant agent to be registered as a "Tool" within another agent's context, managing the lifecycle, state, and security boundaries of these nested sessions.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable registration of "Agent Servers" (like Claude Code) as standard MCP tools.
    *   Manage recursive context inheritance (passing parent intent to child agent).
    *   Implement "Loop Protection" to prevent infinite agent-to-agent call cycles.
    *   Standardize identity propagation (OAuth) through the agent chain.
*   **Non-Goals:**
    *   Building a new LLM agent from scratch.
    *   Managing non-MCP-compliant agent frameworks (handled by the A2A Bridge).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise AI Architect
*   **Primary Goal:** Delegate "Security Audit" of a codebase from a general-purpose Orchestrator Agent to a specialized Security Agent sub-process.
*   **The Happy Path (Tasks):**
    1.  Architect configures a `security-agent` service in MCP Any pointing to a local specialized agent binary.
    2.  The Orchestrator Agent identifies a security task and calls the `security-agent` tool.
    3.  MCP Any intercepts the call, injects parent context headers (session ID, user identity).
    4.  MCP Any starts/attaches to the sub-agent process via Stdio.
    5.  The sub-agent executes, and MCP Any streams results back to the Orchestrator.
    6.  The Orchestrator receives the summarized audit and proceeds.

## 4. Design & Architecture
*   **System Flow:**
    `Orchestrator LLM` -> `MCP Any (Gateway)` -> `Recursive Proxy Middleware` -> `Sub-Agent (e.g., Claude Code)`
*   **APIs / Interfaces:**
    *   New `agent` service type in `config.yaml`.
    *   Standardized headers for recursion depth: `x-mcp-depth`, `x-mcp-parent-session`.
*   **Data Storage/State:**
    *   Session-bound KV store for shared state between parent and child agents.

## 5. Alternatives Considered
*   **Direct Execution:** Letting the parent agent just run the sub-agent as a `bash` command. *Rejected* because it loses context, security controls, and observability.
*   **Flat Toolset:** Manually adding every sub-agent tool to the parent. *Rejected* because it causes context bloat and doesn't allow the sub-agent to use its own optimized internal logic.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Sub-agents inherit a "Restricted Scope" token from the parent. If the parent doesn't have `fs:write` access, the child cannot gain it.
*   **Observability:** Visualizing the agent hierarchy in the UI (Agent Chain Tracer).

## 7. Evolutionary Changelog
*   **2026-03-06:** Initial Document Creation.
