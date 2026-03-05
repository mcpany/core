# Design Doc: Cryptographic Capability Delegation (Auth Chains)
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
Multi-agent swarms frequently use "Sub-agents" for specialized tasks. Currently, sub-agents often either lack permissions or have too many (inheriting the parent's full profile). MCP Any needs a way for parent agents to "sign" a restricted capability grant for a sub-agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a verifiable token protocol for parent-to-subagent permission handoff.
    * Enable granular, time-bound, and intent-scoped access control.
    * Support offline verification of delegated capabilities.
* **Non-Goals:**
    * Creating a new identity provider (this uses existing parent identities).
    * Enforcing tool execution (delegation only authorizes the call).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-agent Swarm Orchestrator (e.g., OpenClaw, CrewAI)
* **Primary Goal:** Securely delegate a subset of parent agent's capabilities to a sub-agent for a specific task.
* **The Happy Path (Tasks):**
    1. Parent agent identifies a sub-task requiring a subset of tools.
    2. Parent agent requests a `DelegationToken` from MCP Any's Auth Engine for specific tools.
    3. MCP Any signs the token with a parent-linked key.
    4. Parent agent passes the `DelegationToken` to the sub-agent.
    5. Sub-agent uses the `DelegationToken` when calling the specified tools.
    6. MCP Any verifies the token's signature and scope before execution.

## 4. Design & Architecture
* **System Flow:**
    `Parent` -> `MCP Any (Auth Engine: auth/delegate)` -> `Signed Token` -> `Sub-agent` -> `MCP Any (JSON-RPC: tools/call + Token)` -> `Policy Engine` -> `Execution`
* **APIs / Interfaces:**
    * `auth/delegate`: New endpoint for generating delegation tokens.
    * `tools/call`: Extended to accept a `delegation_token` header/parameter.
* **Data Storage/State:**
    * Ephemeral storage of active tokens (optional if using JWT/JWS style self-contained tokens).

## 5. Alternatives Considered
* **Dynamic Profile Updates:** Updating the sub-agent's profile on-the-fly. *Rejected* because it's stateful and complex to manage in a distributed swarm.
* **Parent Proxying:** Having all sub-agent calls go through the parent. *Rejected* due to latency and parent context pollution.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Prevent "Capability Escalation" by ensuring the delegation is a strict subset of the parent's existing permissions.
* **Observability:** Audit logs should show the "delegation chain" (Parent -> Sub-agent -> Tool).

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
