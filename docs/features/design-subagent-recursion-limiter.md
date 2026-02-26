# Design Doc: Subagent Recursion Limiter Middleware

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of recursive agent frameworks like OpenClaw, subagents are increasingly spawning their own subagents. This leads to "Context Depth Exhaustion" where the original intent is lost, or the LLM's context window is filled with intermediate reasoning. MCP Any needs to manage this recursion to ensure stability, security, and cost-efficiency.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce a maximum depth for recursive subagent calls.
    *   Provide "Intent-Based Pruning" to pass only relevant context to deep subagents.
    *   Implement "Recursive Scopes" that restrict subagent permissions based on depth.
    *   Inject recursion metadata into tool schemas to inform LLMs of depth limits.
*   **Non-Goals:**
    *   Managing the actual lifecycle of subagents (that's the framework's job).
    *   Solving the General AI alignment problem.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw user).
*   **Primary Goal:** Prevent an autonomous agent from entering an infinite loop of subagent creation that drains API credits.
*   **The Happy Path (Tasks):**
    1.  The user configures a global `max_recursion_depth: 3` in `mcp.yaml`.
    2.  Parent Agent spawns Subagent A (Depth 1).
    3.  Subagent A spawns Subagent B (Depth 2).
    4.  Subagent B spawns Subagent C (Depth 3).
    5.  Subagent C attempts to spawn Subagent D.
    6.  MCP Any intercepts the call, sees the depth is 4, and returns a standardized "Recursion Depth Exceeded" error.
    7.  The chain is gracefully terminated or handled by the parent agents.

## 4. Design & Architecture
*   **System Flow:**
    *   MCP Any injects a `x-mcp-recursion-depth` header into every agent request.
    *   Middleware increments this header on every "spawn" or tool call that indicates a new agent level.
    *   The Policy Engine checks this header against the configured limit.
*   **APIs / Interfaces:**
    *   New Header: `X-MCP-Agent-Depth`.
    *   New Error Code: `MCP_ERROR_RECURSION_LIMIT_EXCEEDED (-32001)`.
*   **Data Storage/State:**
    *   Session-bound recursion counters stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
*   **Trusting the LLM**: Rejected because LLMs are prone to hallucinations and can't reliably self-limit in complex swarms.
*   **Hard-Coded Limits in Frameworks**: Rejected because it's inconsistent across different frameworks (OpenClaw vs. AutoGen). MCP Any provides a universal enforcement layer.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Prevents "Recursion Bomb" denial-of-service attacks where an agent recursively calls expensive tools.
*   **Observability:** All recursion events are logged with their depth, visible in the UI Trace view.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
