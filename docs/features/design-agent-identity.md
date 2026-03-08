# Design Doc: X-MCP-Agent-ID Identity Propagator

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
As agent swarms evolve from single-loop systems to complex, nested hierarchies (e.g., OpenClaw's deterministic sub-agent spawning), the "Identity Crisis" becomes a major security bottleneck. Currently, a tool call from a "Researcher Sub-agent" looks identical to a call from the "Admin Parent Agent." To enforce the Principle of Least Privilege, MCP Any must be able to identify *which* specific agent in a swarm is making a request and what its intended scope is.

## 2. Goals & Non-Goals
* **Goals:**
    * Define a standardized header `X-MCP-Agent-ID` for all tool-call requests.
    * Implement a "JWT-like" Agent Identity Token that includes parent-agent signatures and capability scopes.
    * Ensure identity propagation through multiple layers of agent nesting.
    * Integrate with the Policy Firewall to allow/deny tool calls based on agent identity.
* **Non-Goals:**
    * Providing a global identity provider (IDP) for all AI (focus is on local/session-based identity).
    * Replacing existing authentication (this is *authorization* metadata for agents).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator (e.g., OpenClaw)
* **Primary Goal:** Spawning a sub-agent with access ONLY to a specific directory, even if the parent has full filesystem access.
* **The Happy Path (Tasks):**
    1. Parent agent requests an `IdentityToken` for a sub-agent with scope `fs:read:/tmp/workspace`.
    2. MCP Any issues a signed token bound to the sub-agent's `Agent-ID`.
    3. The sub-agent includes `X-MCP-Agent-ID: [token]` in its tool calls.
    4. The Policy Firewall verifies the token and restricts the `read_file` tool call to the approved directory.

## 4. Design & Architecture
* **System Flow:**
    - **Identity Middleware**: Intercepts JSON-RPC/gRPC calls and extracts the identity header.
    - **Token Validator**: Verifies the signature of the `Agent-ID` token against the session's root key.
    - **Context Injector**: Injects the verified agent identity into the tool execution context for use by the Policy Engine.
* **APIs / Interfaces:**
    - New MCP Tool: `mcp_issue_subagent_identity(scope, duration)`
    - Header: `X-MCP-Agent-ID: <Base64-Token>`
* **Data Storage/State:** Ephemeral session keys stored in memory; no persistent storage required for identity tokens.

## 5. Alternatives Considered
* **Separate Port per Agent**: Running a separate MCP Any instance for every sub-agent. *Rejected* as it is resource-heavy and hard to coordinate.
* **Prefix-based Identity**: Adding `[Agent-Name]` to tool calls. *Rejected* because it is easily spoofed.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is the "Who" in the Zero Trust model (Who, What, Where).
* **Observability:** Trace logs will now show the full "Agent Chain" (e.g., `Admin -> Architect -> Coder`).

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
