# Design Doc: JIT Ephemeral Credential Scoping
**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
Autonomous agents often require long-lived OAuth tokens or API keys to interact with upstream services (e.g., GitHub, Slack, AWS). These secrets are frequently stored in the agent's memory or local configuration, making them vulnerable to "memory dumping" or exfiltration if the agent is compromised. The "OpenClaw Crisis" highlighted how a single compromised subagent could expose the entire organization's credential set. `JIT (Just-In-Time) Ephemeral Credential Scoping` aims to eliminate persistent secrets by providing short-lived, single-use tokens that are minted on-demand for specific tool calls.

## 2. Goals & Non-Goals
* **Goals:**
    * Move away from persistent storage of sensitive upstream credentials in agent-accessible environments.
    * Implement a "Credential Vault" within MCP Any that holds the master secrets.
    * Mint "Ephemeral Tokens" (e.g., scoped JWTs or temporary STS tokens) for each tool execution.
    * Automatically expire or revoke tokens immediately after the tool call completes.
    * Bind tokens to a specific "Intent Scope" (e.g., "Allow `github:create_issue` but deny `github:delete_repo`").
* **Non-Goals:**
    * Replacing existing identity providers (e.g., Okta, Auth0).
    * Managing the lifecycle of the *master* secrets (MCP Any acts as a consumer/proxy, not the primary vault).
    * Providing credentials for non-agentic human users.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Architect configuring agent swarm security.
* **Primary Goal:** Ensure that even if a subagent's execution environment is compromised, the attacker cannot reuse its credentials for unauthorized actions.
* **The Happy Path (Tasks):**
    1. Agent requests to call `slack.post_message`.
    2. MCP Any intercepts the request and identifies the need for a Slack OAuth token.
    3. The `JIT Token Service` retrieves the master Slack secret from the encrypted vault.
    4. It requests a scoped, 60-second token from Slack (or generates a scoped proxy token).
    5. MCP Any injects the ephemeral token into the tool call header and executes the upstream request.
    6. Once the response is received, MCP Any invalidates the token.
    7. If the agent tries to use that same token 5 minutes later, it is rejected by MCP Any.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `MCP Any Core` -> `JIT Token Service` -> `Master Vault (Encrypted)` -> `Scoped Token Minting` -> `Upstream Call`
    1. **Vault Integration**: MCP Any integrates with HashiCorp Vault or AWS Secrets Manager to store master keys securely.
    2. **Scoping Logic**: Maps the requested `tool_id` to a specific required scope (e.g., `repo:status`).
    3. **Minting Engine**: Uses OIDC or protocol-specific JIT mechanisms (like GitHub App Installation Tokens) to create the ephemeral credential.
    4. **Token Injection**: Middleware that dynamically replaces `${EPHEMERAL_TOKEN}` placeholders in config with real values at runtime.
* **APIs / Interfaces:**
    * `internal/auth/jit_service.go`: Logic for token lifecycle management.
    * `Policy.ScopeMapping`: Configuration to define which tools require which scopes.
* **Data Storage/State:**
    * `tokens.db`: Tracking of active ephemeral tokens and their expiration timestamps.
    * `Encrypted Key Store`: Secure local storage for master keys (if not using an external vault).

## 5. Alternatives Considered
* **Credential Masking**: Redacting tokens from logs. Rejected as insufficient, as it doesn't prevent theft from memory.
* **Static Scoped Tokens**: Manually creating hundreds of scoped tokens. Rejected as unmanageable at scale.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Tokens are "use-once-and-destroy." The vault itself is protected by strict hardware-backed attestation (if available).
* **Observability**: Every token minting event is logged with the associated `agent_id` and `intent_context`.

## 7. Evolutionary Changelog
* **2026-03-11:** Initial Document Creation.
