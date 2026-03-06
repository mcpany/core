# Design Doc: Origin-Aware Middleware
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
The "ClawJacked" vulnerability (CVE-2026-2256) demonstrated that local AI agents are vulnerable to hijacking from malicious websites running in the user's browser. Since these websites can make requests to `localhost`, they can silently trigger tool calls (like `Shell` execution) without user interaction. MCP Any must implement a middleware that validates the cryptographic origin of every request to ensure it comes from a trusted agent or user-authorized session.

## 2. Goals & Non-Goals
* **Goals:**
    * Validate the caller's origin for every MCP request.
    * Support binary signature verification for local agent processes.
    * Implement a session-based "Handshake" for web-based agent dashboards.
    * Block all requests that lack a valid, signed Origin Token.
* **Non-Goals:**
    * Replacing existing API Key authentication (this is an additional layer).
    * Protecting against physical access to the local machine.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer running an AI agent swarm locally.
* **Primary Goal:** Prevent a malicious website from executing commands on their machine via MCP Any.
* **The Happy Path (Tasks):**
    1. User starts MCP Any and a trusted agent (e.g., Claude Code).
    2. Claude Code performs a secure handshake with MCP Any, providing its signed binary hash.
    3. MCP Any issues a short-lived `OriginToken` to the agent.
    4. Every subsequent tool call from the agent includes this token.
    5. A malicious website tries to call `tools/call` on `localhost:50050`.
    6. MCP Any rejects the request because it lacks a valid `OriginToken` and doesn't match a trusted binary signature.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: The middleware sits at the edge of the JSON-RPC handler.
    - **Validation**: It checks the `Authorization-Origin` header for a signed JWT or a platform-specific process credential (e.g., `SO_PEERCRED` on Linux).
    - **Policy Engine**: Cross-references the origin with a whitelist of "Approved Agents" and "User Sessions."
* **APIs / Interfaces:**
    - `POST /v1/auth/handshake`: Exchange agent credentials for an `OriginToken`.
    - `GET /v1/auth/origin-status`: Return details about the current request's origin for debugging.
* **Data Storage/State:** In-memory cache of active, authorized `OriginTokens`.

## 5. Alternatives Considered
* **CORS-Only Protection**: Relying on browser CORS. *Rejected* because `curl` and non-browser tools bypass CORS, and some "local-to-local" browser requests can still be dangerous.
* **Prompting for Every Call**: Asking the user to approve every tool call. *Rejected* as it destroys the "autonomous" part of AI agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** This is a core Zero Trust component, ensuring that "Location" (localhost) is not a proxy for "Trust."
* **Observability:** Audit log every rejected origin attempt, including the suspect headers and source IP.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
