# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
In multi-agent swarms (e.g., OpenClaw), a "Lead" agent often spawns "Subagents" to perform specific tasks. Currently, these subagents lose the context, session state, and credentials of the lead agent unless they are manually passed as arguments, which is insecure and leads to "hallucinated state". The Recursive Context Protocol (RCP) standardizes how context is inherited through the MCP chain.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize `_mcp_context` header for passing session state.
    * Enable secure delegation of credentials from parent to child.
    * Provide a mechanism for "Context Filtering" (only pass relevant state to subagents).
* **Non-Goals:**
    * Long-term state persistence (this is handled by the Shared KV Store).
    * Global shared memory across unrelated sessions.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator Developer
* **Primary Goal:** Share a secure API token from a primary research agent to a specialized "Scraper" subagent without exposing the token in the tool arguments.
* **The Happy Path (Tasks):**
    1. Parent agent calls MCP Any with `research_task`.
    2. MCP Any attaches a session-scoped `_mcp_context` containing the token.
    3. Parent agent calls `spawn_subagent` (a tool exposed by MCP Any).
    4. MCP Any propagates the `_mcp_context` to the new subagent session.
    5. The subagent calls `scrape_url`.
    6. The Scraper adapter retrieves the token from the inherited context to authorize the request.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent` --(Call with Context)--> `MCP Any` --(Inject Context)--> `Subagent`
* **APIs / Interfaces:**
    Standardized header: `X-MCP-Context: <base64-encoded-json>`
    The JSON payload contains `session_id`, `scopes`, and `ephemeral_credentials`.
* **Data Storage/State:**
    Context is transient and tied to the JSON-RPC session.

## 5. Alternatives Considered
* **Environment Variable Inheritance**: Rejected as insecure and not portable across remote sessions.
* **Database-backed Context**: Rejected as too high-latency for transient session state.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Context must be cryptographically signed by MCP Any to prevent subagents from escalating their own privileges.
* **Observability**: Traces should include the `context_id` to link parent and child agent calls.

## 7. Evolutionary Changelog
* **2026-02-23**: Initial Document Creation.
