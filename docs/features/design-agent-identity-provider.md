# Design Doc: Agent Identity Provider (AIdP) & NHI Orchestration

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
With the rise of autonomous agents (OpenClaw, Gemini CLI, Claude Code) and coordinating swarms, a critical security gap has emerged: how do agents identify themselves to tools and other agents without sharing static, long-lived user credentials? The "GTG-1002" swarm attack demonstrated that current "all-or-nothing" credential models are vulnerable. MCP Any must evolve to become an **Agent Identity Provider (AIdP)**, issuing task-scoped, Non-Human Identity (NHI) tokens that represent the agent's intent and specific authorized capabilities.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a system for issuing short-lived (task-bound) NHI tokens for agents.
    *   Enable "Identity Bridging" between different agent frameworks (e.g., an OpenClaw agent calling an AutoGen agent via MCP Any).
    *   Support "Task-Scoped Credential Injection," where MCP Any provides ephemeral secrets to tools based on the active NHI token.
    *   Provide a standardized `identity/whoami` and `identity/exchange` MCP extension.
*   **Non-Goals:**
    *   Replacing traditional User Identity Providers (Okta, Auth0, etc.).
    *   Implementing long-term persistent storage for agent secrets (secrets should be ephemeral).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-conscious Enterprise Architect.
*   **Primary Goal:** Allow a specialized research agent to access a private database without giving the agent a permanent database password.
*   **The Happy Path (Tasks):**
    1.  The agent starts a new task and requests an NHI token from MCP Any.
    2.  MCP Any verifies the agent's identity (e.g., via a pre-registered public key) and the user's overarching intent.
    3.  MCP Any issues an NHI token scoped to: `service:postgres:read:table=research_data`.
    4.  The agent calls the `query_database` tool, passing the NHI token.
    5.  MCP Any's middleware intercepts the call, validates the token, and injects an ephemeral DB credential into the tool execution environment.
    6.  The tool executes, and the token/credential expires immediately after the task.

## 4. Design & Architecture
*   **System Flow:**
    - **Token Issuance**: The `IdentityManager` issues JWT-based NHI tokens signed by the MCP Any instance's private key.
    - **Credential Vault**: A secure, in-memory cache of task-scoped credentials mapped to NHI tokens.
    - **Injection Middleware**: Intercepts `tools/call` and populates environment variables/headers with the required secrets.
*   **APIs / Interfaces:**
    - `mcp_any_get_token(intent_id: string) -> token: string`
    - `mcp_any_verify_identity(token: string) -> identity_info: object`
*   **Data Storage/State:** Tokens are stateless (JWT); ephemeral credentials are kept in the `Shared KV Store` (Blackboard) with strict TTLs.

## 5. Alternatives Considered
*   **User-Delegated Static Tokens**: Forcing users to create many manual, scoped tokens. *Rejected* as it doesn't scale and is prone to human error.
*   **Dynamic IAM Role Assumption**: Integrating directly with AWS/GCP/Azure IAM. *Rejected* as a core requirement because MCP Any must work in local/hybrid environments without cloud dependency, though this can be an extension.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of "Zero Trust for Agents." It ensures that even if an agent is compromised, the damage is limited to the current task's scope.
*   **Observability:** All token issuances and credential injections are logged in the `Audit Log`, providing a clear trail for compliance (SOC2/GDPR).

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
