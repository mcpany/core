# Design Doc: Origin-Locked Gateway Middleware
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
With the discovery of the "ClawJacked" vulnerability in OpenClaw, it's clear that local AI agent gateways are vulnerable to Cross-Origin Resource Sharing (CORS) and Origin-based attacks. MCP Any, as a local hub for powerful tools, must ensure that only authorized local applications (e.g., specific IDE extensions, CLI tools) or attested remote nodes can interact with its API and WebSocket endpoints.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement strict Origin verification for all HTTP and WebSocket requests.
    * Provide a mechanism for "attested" local applications to register their origin.
    * Support Unix Domain Sockets (UDS) as a more secure alternative to TCP for local IPC.
    * Log and alert on blocked origin attempts.
* **Non-Goals:**
    * Replacing existing authentication (API keys/Tokens). This is an additional layer of security.
    * Managing network-level firewalls (this is application-level origin trust).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an IDE extension (e.g., VS Code MCP) and a terminal agent.
* **Primary Goal:** Prevent a malicious website visited in a browser from silently calling the `execute_tool` endpoint on the local MCP Any server.
* **The Happy Path (Tasks):**
    1. MCP Any starts and binds to a Unix Domain Socket and/or Localhost with a randomized secret token.
    2. The IDE extension reads the secret token and uses it to perform an initial handshake, registering its unique origin/identifier.
    3. MCP Any issues a session-bound capability token.
    4. Subsequent tool calls include the token and the verified origin.
    5. A malicious website tries to `fetch('http://localhost:3000/execute_tool')`; MCP Any detects a browser origin and rejects it immediately due to missing/mismatched origin attestation.

## 4. Design & Architecture
* **System Flow:**
    `Request -> Origin Verification Middleware -> Token Auth -> Policy Engine -> Tool Execution`
* **APIs / Interfaces:**
    * New `POST /auth/attest`: Initial handshake for local apps to provide their process ID and identity.
    * Middleware: `OriginCheck(next)` which validates `Origin` and `Referer` headers against an allowlist of "Attested Origins."
* **Data Storage/State:**
    * Volatile "Attested Origins" registry in memory.

## 5. Alternatives Considered
* **Strict CORS only**: Rejected because CORS can sometimes be bypassed or misconfigured. Origin + Secret Token + UDS provides defense-in-depth.
* **Manual Allowlist**: Too much friction for users. Automating the handshake for trusted local apps is preferred.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mitigates "Confused Deputy" attacks where a browser is tricked into acting on behalf of the user.
* **Observability:** Audit logs will capture all rejected origins for security forensics.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
