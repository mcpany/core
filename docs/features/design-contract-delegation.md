# Design Doc: Contract-Based Delegation Middleware

**Status:** Draft
**Created:** 2026-03-08

## 1. Context and Scope
As agent swarms (like OpenClaw and CrewAI) grow more complex, simple "allow/deny" permissioning is no longer sufficient. When a parent agent delegates a task to a subagent, it needs to provide a "Capability Contract" that strictly limits the subagent's tool access and data visibility to only what is necessary for that specific task. This prevents subagents from overreaching or being exploited via prompt injection to access unauthorized parent context.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a dynamic middleware that enforces "Capability Contracts" during agent-to-agent (A2A) handoffs.
    *   Allow parent agents to define a subset of their own permissions to be inherited by the subagent.
    *   Provide a standardized "Contract" schema that can be verified by both the parent and the subagent.
    *   Integrate with the `Policy Firewall` for runtime enforcement.
*   **Non-Goals:**
    *   Hard-coding agent logic (the middleware only enforces the contract).
    *   Replacing static RBAC (it complements it with dynamic, task-scoped rules).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Swarm Orchestrator (e.g., OpenClaw).
*   **Primary Goal:** Delegate a "File Analysis" task to a specialized subagent without giving it access to the parent's "Write" or "Delete" tools.
*   **The Happy Path (Tasks):**
    1.  The parent agent initiates a delegation request via MCP Any's A2A Bridge.
    2.  The parent defines a contract: `{"allowed_tools": ["fs:read", "text:analyze"], "expires_in": "5m"}`.
    3.  MCP Any's `ContractMiddleware` intercepts the request and generates a temporary `Contract-Token`.
    4.  The subagent receives the token and the task.
    5.  When the subagent calls `fs:delete`, the `ContractMiddleware` rejects it because it's not in the contract, even if the parent agent has that permission.

## 4. Design & Architecture
*   **System Flow:**
    - **Contract Definition**: Parent agent sends a signed contract along with the delegation request.
    - **Tokenization**: MCP Any validates the signature and issues a scoped, short-lived JWT representing the contract.
    - **Enforcement**: Every tool call made using the `Contract-Token` is checked against the contract's `allowed_tools` and `constraints`.
*   **APIs / Interfaces:**
    - New header: `X-MCP-Contract-Token`.
    - Contract Schema: Standardized JSON-LD or Rego-compatible format.
*   **Data Storage/State:** Tokens are stateless (JWTs), but the `Shared KV Store` may be used to track contract usage and exhaustion.

## 5. Alternatives Considered
*   **Static Permission Sets**: Creating a "Subagent" role. *Rejected* because it's too rigid for dynamic task-based swarms.
*   **Context Filtering**: Only sending specific tool schemas to the subagent. *Rejected* as it doesn't prevent a subagent from guessing the name of a tool it wasn't given.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All contracts are cryptographically signed. Use the "Safe-by-Default" principle—if a tool isn't in the contract, it's denied.
*   **Observability:** Audit logs will record tool calls alongside the contract ID, allowing for "Contract Breach" detection.

## 7. Evolutionary Changelog
*   **2026-03-08:** Initial Document Creation.
