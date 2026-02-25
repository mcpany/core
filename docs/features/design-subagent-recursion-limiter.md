# Design Doc: Subagent Recursion Limiter Middleware

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of recursive agent frameworks like OpenClaw, subagents are increasingly spawning their own sub-subagents. This leads to "Context Depth Exhaustion" and potential infinite loops where agents pass tasks back and forth. MCP Any, as the universal gateway, is uniquely positioned to enforce depth limits and prune context to ensure reliability and cost-efficiency.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce a maximum recursion depth for agent sessions.
    *   Prune "stale" or "intermediate" context as agents descend into deeper sub-tasks.
    *   Provide a standard header for tracking "Origin Intent" across depths.
*   **Non-Goals:**
    *   Determining the *content* of the agent's reasoning (that's the LLM's job).
    *   Managing local process lifecycles (this only handles the MCP protocol layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator
*   **Primary Goal:** Prevent a runaway subagent from consuming thousands of tokens or entering an infinite loop.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator starts a session with `X-MCP-Max-Depth: 3`.
    2.  Parent Agent calls a subagent (Depth 1).
    3.  Subagent calls another subagent (Depth 2).
    4.  The leaf subagent attempts a recursive call that would exceed Depth 3.
    5.  MCP Any blocks the request with a `429 Too Many Requests` or a specialized JSON-RPC error indicating "Recursion Limit Exceeded".

## 4. Design & Architecture
*   **System Flow:**
    1.  **Request Ingestion**: Middleware checks for the `X-MCP-Session-ID` and `X-MCP-Depth` headers.
    2.  **Depth Validation**: If `X-MCP-Depth` >= `X-MCP-Max-Depth`, the request is rejected.
    3.  **Context Pruning**: If allowed, the middleware applies a "Context Squeezer" that keeps only the `Origin-Intent` and the immediate parent's prompt.
    4.  **Header Injection**: Middleware increments `X-MCP-Depth` and passes the request upstream.
*   **APIs / Interfaces:**
    *   New Header: `X-MCP-Depth` (int)
    *   New Header: `X-MCP-Max-Depth` (int, default 5)
    *   New Header: `X-MCP-Origin-Intent` (string)
*   **Data Storage/State:**
    *   Session state stored in the existing Shared KV Store (Blackboard) to track depth across distributed calls.

## 5. Alternatives Considered
*   **Client-side enforcement**: Rejected because agents are autonomous and might bypass or forget to implement limits.
*   **LLM-based monitoring**: Rejected due to high latency and cost; this needs to be a hard infrastructure-level constraint.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Prevents Resource Exhaustion attacks (Denial of Wallet).
*   **Observability:** Each pruned request or blocked call is logged to the `mcp-audit-log`.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
