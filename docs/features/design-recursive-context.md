# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
In agent swarm architectures, a primary agent often delegates tasks to subagents. Currently, these subagents lose the critical context of the parent agent (e.g., authorization headers, session IDs, trace IDs, and high-level goals). The Recursive Context Protocol standardizes how this metadata is propagated through MCP Any.

## 2. Goals & Non-Goals
* **Goals:**
    * Define a standardized set of MCP metadata headers for context propagation.
    * Automatically inject parent context into subagent tool calls.
    * Support trace ID propagation for end-to-end observability across agent chains.
* **Non-Goals:**
    * Managing the lifecycle of subagents (orchestrators like OpenClaw handle this).
    * Providing a persistent long-term memory (covered by Shared KV Store).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Orchestrator / Swarm Developer
* **Primary Goal:** Ensure a subagent's actions are correctly attributed to the parent session and inherit its security constraints.
* **The Happy Path (Tasks):**
    1. Parent agent calls a tool that triggers a subagent.
    2. Parent agent includes `X-MCP-Context` header in the tool call.
    3. MCP Any receives the call, extracts the context, and stores it in the request context.
    4. When the subagent is invoked via MCP Any, the server automatically injects the inherited headers.
    5. The entire chain is visible in the observability dashboard under a single Trace ID.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent -> [X-MCP-Context] -> MCP Any -> Subagent -> [Injected Context] -> Upstream`
* **APIs / Interfaces:**
    * Extension of the MCP `meta` object in JSON-RPC requests.
    * Middleware to extract/inject headers.
* **Data Storage/State:**
    * Context is passed in-band with requests.

## 5. Alternatives Considered
* **Manual Context Passing:** Prone to developer error and lacks standardization across different agent frameworks.
* **Out-of-band State Store:** Adds latency and complexity; in-band headers are more efficient for transient context.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Ensures that subagents cannot escalate privileges beyond their parent.
* **Observability:** Critical for debugging multi-agent interactions.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
