# Design Doc: Recursive Context Protocol
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As agent swarms become more complex (e.g., OpenClaw subagents), there is a need for a standardized way to pass security context, authentication tokens, and session state from a parent agent to its subagents. Currently, this is often handled in a fragmented, insecure way, or not handled at all, leading to subagents losing access to necessary tools or data.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize MCP headers for context inheritance.
    * Allow subagents to securely "bootstrap" from parent context.
    * Support propagation of OAuth2/OIDC tokens across agent boundaries.
    * Enable "Trace ID" propagation for multi-agent observability.
* **Non-Goals:**
    * Implementing the subagent orchestration logic itself.
    * Storing large amounts of state in headers (state should be in the Blackboard tool).

## 3. Critical User Journey (CUJ)
* **User Persona:** Agent Swarm Developer
* **Primary Goal:** A "Lead Agent" spawns a "Researcher Subagent" that needs to access a private GitHub repository using the Lead Agent's credentials.
* **The Happy Path (Tasks):**
    1. Lead Agent receives an MCP request with an `Authorization` header.
    2. Lead Agent calls the `spawn_subagent` tool provided by MCP Any.
    3. MCP Any injects the `X-MCP-Context-Parent` header into the subagent's environment.
    4. Subagent makes a tool call to `github/read_repo`.
    5. MCP Any's Recursive Context Middleware extracts the credentials from the parent context and applies them to the subagent's request.
    6. Subagent successfully reads the repo without the user having to re-authenticate.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Client->>ParentAgent: tools/call (with Auth)
        ParentAgent->>MCPAny: tools/call (spawn_subagent)
        MCPAny->>SubAgent: Start (with Context Headers)
        SubAgent->>MCPAny: tools/call (github/read_repo)
        MCPAny->>MCPAny: Recursive Context Middleware (Inject Auth)
        MCPAny->>GitHub: API Request (with Auth)
    ```
* **APIs / Interfaces:**
    * `X-MCP-Context-ID`: Unique session identifier for the swarm.
    * `X-MCP-Context-Parent`: ID of the parent agent.
    * `X-MCP-Context-Scope`: Restricted scope for the subagent.

## 5. Alternatives Considered
* **Manual Token Passing**: Rejected as insecure and error-prone.
* **Global Session State**: Rejected due to scalability concerns and lack of granular control over subagent permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Context propagation must include scope restriction. A parent should be able to grant only a subset of its permissions to a child.
* **Observability:** Trace IDs must be preserved across the recursion to allow for full-swarm debugging in the Trace View.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
