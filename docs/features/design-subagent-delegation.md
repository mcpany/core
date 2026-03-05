# Design Doc: Subagent Delegation Protocol
**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As AI agents evolve from monolithic entities into hierarchical swarms (e.g., OpenClaw's Multi-Agent mode), a critical security and operational gap has emerged: the "Delegation Problem." Currently, when a parent agent spawns a subagent, it either gives it full access to its own environment (violating the principle of least privilege) or requires complex, manual configuration of permissions.

MCP Any's Subagent Delegation Protocol (SDP) aims to automate the secure "handoff" of tools, state, and context from a parent to a subagent. It ensures that subagents are confined to a specific "intent-scope" while maintaining the necessary context to perform their specialized tasks.

## 2. Goals & Non-Goals
* **Goals:**
    * Automatically generate restricted session tokens for subagents.
    * Enable "State Slicing" where only relevant parts of the parent's context are inherited.
    * Provide a standardized "Spawn" tool that any agent can call to create a governed subagent.
    * Ensure cryptographic binding between a subagent's token and its specific task.
* **Non-Goals:**
    * Managing the LLM lifecycle (spawning the actual process). SDP manages the *permissions* and *context bridge*.
    * Solving general multi-agent communication (handled by the A2A Bridge).

## 3. Critical User Journey (CUJ)
* **User Persona:** Senior AI Architect / Swarm Orchestrator
* **Primary Goal:** Securely delegate a "Code Review" task to a specialized subagent without giving it access to the production deployment tools.
* **The Happy Path (Tasks):**
    1. Parent Agent identifies a need for a subagent.
    2. Parent Agent calls `mcp_any.delegate(task="review_code", tools=["git_read", "file_read"], context_keys=["diff_v1_v2"])`.
    3. MCP Any SDP generates a short-lived `SubagentToken` and a virtual tool registry for that token.
    4. Parent Agent spawns the subagent, passing the `SubagentToken`.
    5. Subagent attempts to call `deploy_to_prod`; MCP Any denies the request based on the restricted SDP scope.
    6. Subagent finishes and returns result; SDP invalidates the token.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>SDP Middleware: Request Delegation (Scope + Context)
        SDP Middleware->>PolicyEngine: Validate Delegation Request
        PolicyEngine-->>SDP Middleware: Approved
        SDP Middleware->>TokenStore: Generate Scoped Token
        SDP Middleware-->>ParentAgent: Return SubagentToken + ProxyURL
        ParentAgent->>SubAgent: Spawn with Token
        SubAgent->>SDP Middleware: Call Tool (with SubagentToken)
        SDP Middleware->>SDP Middleware: Enforce Scope
        SDP Middleware->>MCP Server: Execute Tool
        MCP Server-->>SubAgent: Result
    ```
* **APIs / Interfaces:**
    * `rpc delegate(scope: Scope, context_slice: State) returns (SubagentSession)`
    * `rpc revoke(session_id: string)`
* **Data Storage/State:**
    * Transient `DelegationSession` stored in-memory (or Redis for distributed nodes), mapping tokens to specific tool-allowlists and filtered context slices.

## 5. Alternatives Considered
* **Static Configs**: Requires pre-defining every possible subagent type. Too rigid for dynamic agent reasoning.
* **Manual Token Passing**: Error-prone and risks "Token Leakage" where a subagent could reuse a parent's token.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are bound to the specific `TaskID`. Even if stolen, they cannot be used for tasks outside the original delegation scope.
* **Observability:** Every delegated tool call is logged with the `ParentID -> SubagentID` lineage for auditability.

## 7. Evolutionary Changelog
* **2026-03-03:** Initial Document Creation.
