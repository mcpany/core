# Design Doc: Agent Identity Provider (OIDC Bridge)
**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
Currently, AI agents often operate using "Shadow Identities"—they share the same API keys or session tokens as the human user who launched them. This makes it impossible for security teams to distinguish between a human-initiated action and an autonomous agent-initiated action during an audit. Furthermore, as agents interact with enterprise systems, they need their own cryptographically verifiable identities to support "On-Behalf-Of" (OBO) authentication flows.

MCP Any will implement an OIDC Bridge that issues temporary, short-lived identities to specific agent sessions, allowing for granular permissioning and perfect auditability.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue OIDC-compliant identity tokens (JWTs) to agent sessions.
    * Support "Agent-on-Behalf-of-User" flows where an agent's permissions are a subset of the user's.
    * Provide a standard "WhoAmI" tool for agents to discover their own identity and scope.
    * Integrate with existing OIDC providers (Okta, Auth0, Google) for user-to-agent identity delegation.
* **Non-Goals:**
    * Replacing the primary user identity provider.
    * Managing human user passwords.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Audit all file deletions to determine if they were done by the "CodeReviewer" agent or the "LeadDev" human.
* **The Happy Path (Tasks):**
    1. The architect configures the `AgentIdentity` provider in MCP Any.
    2. A human user ("LeadDev") authenticates and spawns a "RefactorAgent."
    3. MCP Any issues a scoped JWT to "RefactorAgent" with a `sub` claim linked to "LeadDev" but a unique `agent_id`.
    4. The agent calls a "DeleteFile" tool.
    5. The tool execution log records: `User: LeadDev, Agent: RefactorAgent, Action: Delete, File: /src/old.go`.
    6. The architect reviews the logs and sees exactly which agent performed the action.

## 4. Design & Architecture
* **System Flow:**
    `Human -> MCP Any (Auth) -> Agent Session Start -> Issue Agent JWT -> Tool Call with Agent JWT -> Audit Log`
* **APIs / Interfaces:**
    * `/auth/agent/token`: Endpoint to exchange user session for agent session token.
    * `mcpserver` update to include `IdentityContext` in all tool call metadata.
* **Data Storage/State:**
    * RSA/ECDSA keys for signing Agent JWTs (managed by MCP Any).
    * Mapping of active agent sessions to parent user identities in memory or Redis.

## 5. Alternatives Considered
* **Static API Keys per Agent**: Rejected because it doesn't support the dynamic nature of agent swarms and is a nightmare to manage at scale.
* **User Credential Sharing**: Rejected as it violates the principle of least privilege and destroys audit trails.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Agent tokens are extremely short-lived (e.g., 5-15 minutes) and limited to the specific tools required for their task.
* **Observability:** All JWT issuance and validation events are logged.

## 7. Evolutionary Changelog
* **2026-03-04:** Initial Document Creation.
