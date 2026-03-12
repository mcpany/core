# Design Doc: Ephemeral Agent Identity (Token Exchange)
**Status:** Draft
**Created:** 2026-03-12

## 1. Context and Scope
In multi-agent swarms, subagents often require credentials to access tools or external APIs. Distributing long-lived parent credentials to every subagent creates a massive "Blast Radius" in case of a subagent compromise. Furthermore, identifying *which* subagent made a specific call is difficult without a granular identity system. MCP Any needs a way to issue short-lived, task-scoped, and hardware-attested identities to subagents.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Token Exchange Service" that converts parent credentials into ephemeral JWTs.
    * Support task-scoped permissions (e.g., this token is only valid for 10 minutes and for tool `fs:read`).
    * Integrate hardware-rooted attestation (TPM/HSM) into the token signing process.
    * Provide mutual TLS or similar cryptographic proof of subagent identity.
* **Non-Goals:**
    * Replacing existing OIDC/OAuth providers for human users.
    * Managing the full lifecycle of external API keys (focus is on internal agent identity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw).
* **Primary Goal:** Securely delegate limited permissions to a specialized "Researcher" subagent.
* **The Happy Path (Tasks):**
    1. Parent Agent requests an ephemeral token for a subagent from MCP Any.
    2. MCP Any validates the Parent's hardware identity.
    3. MCP Any issues a 15-minute JWT scoped to specific tools.
    4. Subagent uses the JWT to call tools via MCP Any.
    5. MCP Any verifies the JWT and the subagent's hardware signature before executing the tool.

## 4. Design & Architecture
* **System Flow:**
    `Parent Agent` -> `Token Exchange API` -> `HSM/TPM (Signing)` -> `JWT`
    `Subagent` -> `Tool Call (with JWT)` -> `MCP Any (Validation)` -> `Upstream Tool`
* **APIs / Interfaces:**
    * `POST /v1/identity/exchange`: Parent requests a subagent token.
    * `GET /v1/identity/verify`: Middleware endpoint to verify an incoming subagent JWT.
* **Data Storage/State:**
    * `token_registry.db`: Tracks active ephemeral tokens and their scopes for revocation.

## 5. Alternatives Considered
* **Credential Injection**: Passing raw API keys to subagents via env vars. Rejected due to high exfiltration risk.
* **Static Subagent Keys**: Giving every subagent a permanent key. Rejected due to management overhead and risk.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Tokens are hardware-rooted. If the TPM chip is not present or the signature doesn't match, the token is rejected.
* **Observability**: All token exchanges and usages are logged in the `Identity Audit Log`, allowing for precise "Agent Provenance" tracking.

## 7. Evolutionary Changelog
* **2026-03-12:** Initial Document Creation.
