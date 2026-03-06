# Design Doc: Agent Identity & Attestation (AIA)
**Status:** Draft
**Created:** 2026-03-02

## 1. Context and Scope
As AI agent swarms (like OpenClaw and CrewAI) become more prevalent, the lack of robust identity management for individual agents has become a critical security flaw. The Moltbook breach demonstrated that long-lived API tokens shared across a swarm are "honey pots" for attackers. Claude 4.6's benchmark performance (86.8%) proves that multi-agent systems are the future, and they require specialized infrastructure. MCP Any must provide a way to issue, verify, and rotate identities for every agent and subagent within its ecosystem.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Issue short-lived, session-bound cryptographic identities (JWT or Macaroons) to agents.
    *   Enable "Identity-Aware" tool permissions (e.g., only the "Researcher" agent can call the search tool).
    *   Provide an attestation mechanism for agent frameworks to prove they are running in a secure environment.
    *   Automatic token rotation and revocation.
    *   Support for multi-protocol identity propagation (MCP, A2A, UCP).
*   **Non-Goals:**
    *   Replacing human OIDC/OAuth (this is for *Agent* identity, not user identity).
    *   Managing LLM provider API keys (AIA focuses on the *Agent's* identity within the MCP Any bus).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Multi-Agent Swarm Orchestrator (e.g., a CrewAI script).
*   **Primary Goal:** Initialize a swarm where each agent has distinct permissions and cannot leak the other agents' credentials.
*   **The Happy Path (Tasks):**
    1.  The Orchestrator requests an "Initial Identity" from MCP Any using its own master credential.
    2.  As the Orchestrator spawns subagents, it requests "Scoped Identities" for each (e.g., "Agent: Researcher", "Scope: tool:web_search").
    3.  Each subagent uses its scoped token to call tools via MCP Any.
    4.  MCP Any verifies the token and the scope before proxying the tool call.
    5.  When the session ends, all identities are automatically invalidated.

## 4. Design & Architecture
*   **System Flow:**
    - **Identity Provider (IdP) Middleware**: Internal service within MCP Any that handles token issuance.
    - **Session Vault**: Temporary storage for active agent sessions and their associated public keys.
    - **Verification Loop**: Every tool call is intercepted by the AIA middleware to validate the `Authorization: Agent <token>` header.
*   **APIs / Interfaces:**
    - `POST /v1/identity/issue`: Exchange master credentials for scoped agent tokens.
    - `GET /v1/identity/verify`: Internal endpoint for tool adapters to check identity.
    - Metadata extension for tool calls: `_mcp_agent_id: "agent-abc-123"`
*   **Data Storage/State:** Tokens are ephemeral and stored in-memory (backed by the **Audited Blackboard** if persistence is needed across restarts).

## 5. Alternatives Considered
*   **Standard OAuth2**: Too heavy for local, high-frequency agent handoffs.
*   **Simple API Keys**: Hard to scope and rotate without significant management overhead.
*   **SPIFFE/Spire**: Excellent for service identity but currently lacks the "agentic context" (e.g., "Task ID") needed for AI swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Implements the principle of least privilege at the agent level.
*   **Observability:** AIA logs will provide a detailed audit trail of *which agent* called *which tool* and *why* (if intent is provided).
*   **Interoperability**: Ensures that identities issued by MCP Any are valid across A2A and UCP boundaries.

## 7. Evolutionary Changelog
*   **2026-03-02:** Initial Document Creation.
