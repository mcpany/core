# Design Doc: Cost-Aware Tool Routing & Semantic Caching

**Status:** Draft
**Created:** 2026-03-07

## 1. Context and Scope
As agentic workflows scale, "Tool Call Inflation" (the tendency of LLMs to make redundant or unnecessarily expensive tool calls) is becoming a primary blocker for enterprise adoption. Claude Code and other CLI agents often consume significant token budgets in a single session. MCP Any, as the universal gateway, is uniquely positioned to intercept, estimate, and optimize these calls.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Estimate the cost of tool calls before execution.
    *   Enforce budget-based routing policies (e.g., "Block tools costing >$0.05 without approval").
    *   Provide "Semantic Memory" by caching tool outputs and serving them to agents via similarity search.
    *   Reduce redundant upstream API calls.
*   **Non-Goals:**
    *   Accurately predicting variable costs of dynamic upstreams (only estimates are provided).
    *   Replacing the LLM's own context window management.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Platform Engineer.
*   **Primary Goal:** Prevent " runaway" agents from consuming the entire monthly API budget in a single day.
*   **The Happy Path (Tasks):**
    1.  Engineer configures a `CostPolicy` in MCP Any: `max_cost_per_session: $10.00`.
    2.  An agent calls a `generate_large_report` tool multiple times.
    3.  MCP Any's `CostTelemetryMiddleware` estimates the cost.
    4.  If the budget is exceeded, the middleware intercepts the call and returns a `ResourceExhausted` error with a suggestion to use the Semantic Cache.
    5.  The agent then queries the Semantic Cache for previous report versions, avoiding a new expensive call.

## 4. Design & Architecture
*   **System Flow:**
    - **Cost Estimation**: The `TelemetryMiddleware` uses metadata (injected into tool schemas) to calculate estimated costs.
    - **Semantic Indexing**: Tool outputs are passed through an embedding model (e.g., local BERT or OpenAI `text-embedding-3-small`) and stored in the Blackboard's vector extension.
    - **Similarity Search**: When a tool is called, the middleware first checks the Semantic Cache for high-similarity historical results.
*   **APIs / Interfaces:**
    - **Cost Headers**: `X-MCP-Estimated-Cost` and `X-MCP-Session-Budget`.
    - **Semantic Search Tool**: `query_tool_memory(query, tool_name)`.
*   **Data Storage/State:** Blackboard (SQLite) with `sqlite-vss` for embedding storage.

## 5. Alternatives Considered
*   **LLM-Side Caching**: Letting the model handle memory. *Rejected* because it increases context bloat and doesn't solve the cross-agent cost problem.
*   **Upstream Caching (Redis)**: Traditional key-based caching. *Rejected* because agents rarely send the exact same arguments twice; similarity-based search is required.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Semantic Cache must respect the same RBAC/DLP policies as the original tools. An agent cannot retrieve a cached result for a tool it doesn't have permission to call.
*   **Observability:** Expose real-time budget tracking and cache hit/miss ratios in the UI.

## 7. Evolutionary Changelog
*   **2026-03-07:** Initial Document Creation.
