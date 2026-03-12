# Design Doc: Recursive Depth-Limit Middleware
**Status:** Draft
**Created:** 2026-03-18

## 1. Context and Scope
As agent swarms become more complex, subagents often call other agents or tools that trigger further agentic behavior. This can lead to the "Spiral of Death" (recursive loops) where agents call each other indefinitely, exhausting tokens and compute resources. MCP Any needs a native mechanism to monitor and cap the depth of tool-calling graphs.

## 2. Goals & Non-Goals
* **Goals:**
    * Track call depth across different MCP servers and agent frameworks.
    * Detect circular dependencies in tool-to-tool calls.
    * Provide a "Circuit Breaker" that halts execution when a depth limit is reached.
* **Non-Goals:**
    * Automatically resolving loops (requires LLM reasoning).
    * Restricting parallel tool calls (only vertical recursion).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Orchestrator
* **Primary Goal:** Prevent a buggy or compromised agent from triggering an infinite loop of paid tool calls.
* **The Happy Path (Tasks):**
    1. Parent Agent calls Subagent A.
    2. Subagent A calls Subagent B (Depth: 2).
    3. Subagent B accidentally calls Parent Agent (Detection: Loop).
    4. MCP Any Middleware detects the loop or depth limit.
    5. Middleware returns a standardized `RECURSION_LIMIT_EXCEEDED` error to the caller.
    6. Orchestrator receives the alert and halts the session.

## 4. Design & Architecture
* **System Flow:**
    `Tool Request` -> `Depth Monitor` -> `Call-Graph Update` -> `Limit Check` -> `Execution`
* **APIs / Interfaces:**
    * `X-MCP-Depth` Header: Standardized header for propagating depth across nodes.
    * `X-MCP-Session-ID`: Correlating calls within a single swarm "intent."
* **Data Storage/State:**
    * In-memory graph of active call chains per session.
    * Global depth limit (default: 5) and per-tool overrides.

## 5. Alternatives Considered
* **Time-based TTL**: Rejected because complex tasks might naturally take a long time without being recursive.
* **Token-budget Capping**: Effective for cost but doesn't prevent resource exhaustion on the gateway itself.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevents "Resource Exhaustion" (DoS) attacks from rogue subagents.
* **Observability:** Waterfall visualization of call graphs in the UI.

## 7. Evolutionary Changelog
* **2026-03-18:** Initial Document Creation.
