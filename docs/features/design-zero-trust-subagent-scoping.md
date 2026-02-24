# Design Doc: Zero-Trust Subagent Scoping

**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
As AI agent architectures evolve from monolithic entities to complex swarms of specialized subagents, the security perimeter must become more granular. Currently, if a parent agent has access to a tool, any subagent it spawns often inherits that same level of access. This creates a significant risk of lateral movement if a specialized subagent (e.g., a "code refactorer") is compromised or hallucinates a dangerous command. MCP Any needs to provide a mechanism to restrict subagents to the minimum set of capabilities required for their specific task.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement capability-based tokens (macaroons or scoped JWTs) for subagent tool calls.
    *   Enable "Intent-Aware" scoping where a subagent's permissions are bound to a specific sub-task description.
    *   Provide a middleware that validates these scoped tokens before allowing a tool execution.
*   **Non-Goals:**
    *   Implementing a full identity provider (IdP).
    *   Managing user-level authentication (this design focuses on agent-to-agent/subagent security).

## 3. Critical User Journey (CUJ)
*   **User Persona:** AI Swarm Orchestrator
*   **Primary Goal:** Delegate a "File Read" task to a subagent without giving it "File Write" or "Network" access.
*   **The Happy Path (Tasks):**
    1.  The Parent Agent receives a high-level goal.
    2.  The Parent Agent requests a scoped token from MCP Any for a subagent, specifying the scope (e.g., `fs:read:/docs/*`).
    3.  MCP Any issues a cryptographically signed token bound to that scope.
    4.  The Subagent attempts to call a "File Read" tool using the token.
    5.  MCP Any Middleware verifies the token and allows the call.
    6.  The Subagent attempts to call a "File Write" tool; MCP Any Middleware rejects the call because it's out of scope.

## 4. Design & Architecture
*   **System Flow:**
    - Parent Agent -> MCP Any (Token Minting API) -> Scoped Token
    - Subagent -> MCP Any (Tool Execution API + Scoped Token) -> Policy Engine -> Upstream Tool
*   **APIs / Interfaces:**
    - `POST /v1/auth/mint-scope`: Accepts a parent token and a requested scope; returns a scoped subagent token.
    - Headers: `X-MCP-Subagent-Scope: <token>`
*   **Data Storage/State:**
    - Tokens are stateless and self-describing (signed by MCP Any).
    - Scope definitions follow a standardized URI-like format (e.g., `service:tool:resource:action`).

## 5. Alternatives Considered
*   **Implicit Scoping via Prompting**: Rejected because LLMs cannot be trusted to self-police their permissions (vulnerable to prompt injection).
*   **Dynamic Config Reloading per Subagent**: Rejected due to high latency and complexity in managing thousands of ephemeral configurations.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust)**: All tool calls are denied by default unless a valid, scoped token is present. Tokens are short-lived.
*   **Observability**: All scoped token issuances and rejections are logged in the Audit Trail, including the parent/child relationship.

## 7. Evolutionary Changelog
*   **2026-02-24**: Initial Document Creation.
