# Design Doc: Identity-Bound Tooling (IBT)
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
Autonomous agents often require access to sensitive API keys and credentials to perform their tasks. However, if an agent is compromised (e.g., via prompt injection), these credentials can be exfiltrated and used maliciously. Identity-Bound Tooling (IBT) ensures that credentials are never directly exposed to the agent. Instead, MCP Any manages the credentials and only injects them into tool calls that are cryptographically bound to a verified, short-lived user session.

## 2. Goals & Non-Goals
* **Goals:**
    * Abstract credentials away from the agent runtime.
    * Bind tool execution to a live, user-attested session token.
    * Prevent "Out-of-Band" usage of credentials if the agent's state is exfiltrated.
    * Implement a "Sensitive Sink" to redact secrets from tool responses.
* **Non-Goals:**
    * Implementing a full Identity Provider (IdP); IBT should integrate with existing OIDC/MFA providers.
    * Preventing all forms of agent misuse (focus is on credential protection).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Developer using a shared agent swarm.
* **Primary Goal:** Ensure that even if one of the subagents is compromised, the corporate GitHub Token cannot be exfiltrated.
* **The Happy Path (Tasks):**
    1. User starts MCP Any and authenticates via OIDC.
    2. MCP Any generates a short-lived Session Key.
    3. Agent requests a `list_repositories` tool call.
    4. MCP Any intercepts the call, verifies the Session Key is valid.
    5. MCP Any retrieves the GitHub Token from a secure vault (not visible to agent).
    6. MCP Any executes the tool call, injecting the token into the header.
    7. MCP Any redacts any potential secrets from the GitHub response before returning it to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent (No Credentials)` -> `MCP Any (IBT Middleware)` -> `Upstream API (with Injected Token)`
    1. **Session Management**: IBT Middleware tracks active user sessions and their associated cryptographic material.
    2. **Credential Injection**: Tools are configured with "Secret Placeholders" (e.g., `${IBT_GITHUB_TOKEN}`). These are resolved only at the moment of the upstream request.
    3. **Cryptographic Binding**: Each tool call must include a JWS (JSON Web Signature) signed by the client/user, proving the call originated from an authorized session.
* **APIs / Interfaces:**
    * `MCP Request Header: X-IBT-Session-Token`: Carries the cryptographic proof.
    * `Internal Interface: CredentialProvider`: Pluggable interface for retrieving secrets (Vault, Env, AWS SM).
* **Data Storage/State:**
    * `sessions`: In-memory LRU cache (or Redis) for active session tokens and their TTL.

## 5. Alternatives Considered
* **Agent-level Secret Management**: Rejected because the agent's memory space is considered insecure once a prompt injection occurs.
* **IP Whitelisting**: Insufficient for multi-tenant or cloud-based agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Credentials have a "Zero TTL" in the agent's context. They exist only in the wire-transfer to the upstream API.
* **Observability**: Tool calls are logged with the `SessionID` (but without the secrets) for auditability.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
