# Design Doc: Zero-Trust Subagent Scoping

**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
As agent swarms become more complex, parent agents frequently spawn specialized subagents to handle specific tasks. Currently, these subagents often inherit the full permission set of the parent, violating the principle of least privilege and creating a massive attack surface if a subagent is compromised (e.g., via "Toxic Flow" or prompt injection).

MCP Any needs to provide a mechanism to "scope" permissions for subagents dynamically, ensuring they only have access to the specific tools and resources required for their assigned task.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a capability-based token system for subagent session isolation.
    *   Enable parent agents to define a "restricted toolset" when handshaking with a subagent.
    *   Provide cryptographic proof of scope during tool execution.
*   **Non-Goals:**
    *   Defining the internal logic of the subagents themselves.
    *   Replacing global policy firewalls (this is an additional layer).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Share secure context between 3 agents without exposing local env vars or sensitive tools to untrusted subagents.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator initializes a new subagent session via MCP Any.
    2.  The Orchestrator specifies a list of allowed tool patterns (e.g., `fs:read:/tmp/*`, `weather:*`).
    3.  MCP Any generates a "Scope-Bound Token" for this subagent.
    4.  The subagent attempts to call a forbidden tool (e.g., `fs:write:/etc/passwd`).
    5.  MCP Any Policy Engine rejects the call based on the token's scope, even if the parent agent has permission.

## 4. Design & Architecture
*   **System Flow:**
    Parent Agent -> Request Session(Scope) -> MCP Any (Issuer) -> Scope-Bound Token -> Subagent -> Tool Call(Token) -> MCP Any (Verifier) -> Policy Engine -> Upstream MCP Server.
*   **APIs / Interfaces:**
    *   `mcp_create_scoped_session(parent_token, scope_definition) -> subagent_token`
    *   `scope_definition` follows a CEL (Common Expression Language) or Rego-like pattern for tool/resource matching.
*   **Data Storage/State:**
    Scoped sessions are stored in the transient `Shared KV Store` (Blackboard) and expire with the agent session.

## 5. Alternatives Considered
*   **Standard OIDC Scopes:** Rejected because they are too coarse-grained for the dynamic, resource-specific needs of AI agents.
*   **OS-level Sandboxing:** Useful but doesn't solve the problem of granular tool access within the MCP protocol itself.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Tokens must be short-lived and cryptographically bound to the subagent's session ID to prevent token theft.
*   **Observability:** All scope violations must be logged with "Intent Attribution" to help developers debug swarm behavior.

## 7. Evolutionary Changelog
*   **2026-02-24:** Initial Document Creation.
