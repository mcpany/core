# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
In multi-agent swarms, a "parent" agent often delegates tasks to "subagents." Currently, these subagents lose the authentication context or session history of the parent, leading to "Auth Fragmentation." The Recursive Context Protocol standardizes how MCP Any propagates context across nested agent calls.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Automatically propagate `Authorization` and `X-Session-ID` headers to child tools.
    *   Support context "narrowing" (e.g., a child agent gets a subset of parent permissions).
    *   Compatible with OpenClaw and AutoGen swarm patterns.
*   **Non-Goals:**
    *   Providing a permanent state store (handled by Agentic Blackboard).
    *   Automating agent reasoning about the context.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars.
*   **The Happy Path (Tasks):**
    1.  Parent Agent initiates a session with MCP Any.
    2.  Parent calls a tool that spawns a Subagent.
    3.  MCP Any injects the parent's session metadata into the Subagent's environment.
    4.  Subagent calls its own tools using the inherited credentials seamlessly.

## 4. Design & Architecture
*   **System Flow:**
    *   Recursive Context is handled via MCP `metadata` and custom JSON-RPC headers.
    *   `parent_session_id` is tracked in the server-side session registry.
*   **APIs / Interfaces:**
    *   `mcp_context` object added to tool call metadata.
*   **Data Storage/State:**
    *   Ephemeral session-bound state stored in memory (or Redis for distributed setups).

## 5. Alternatives Considered
*   **Manual Context Passing:** Rejected as it places too much burden on the LLM and increases token usage.
*   **Global Auth:** Rejected as it violates the principle of least privilege.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Context propagation is strictly opt-in per service. Tokens are never logged.
*   **Observability:** Tracing IDs are propagated along with the context for end-to-end swarm debugging.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
