# Design Doc: CSRF-Protected Local Gateway
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Local MCP gateways typically bind to `localhost` and often omit authentication under the assumption that they are only accessible to the local user. However, the OpenClaw RCE (CVE-2026-25253) proved that a malicious website can use a user's browser as a "proxy" to send unauthorized requests to these local ports via Cross-Site Request Forgery (CSRF).

MCP Any needs a robust mechanism to verify that requests to its local API originate from trusted clients (e.g., local IDEs, CLI tools, or authorized web dashboards) and not from rogue browser tabs.

## 2. Goals & Non-Goals
* **Goals:**
    *   Verify the `Origin` and `Host` of all incoming HTTP/WebSocket requests.
    *   Implement a cryptographic handshake (Nonce-based) for non-transparent clients.
    *   Protect against "One-Click RCE" patterns where a simple link click triggers a tool call.
    *   Ensure Zero-Trust even on the local loopback interface.
* **Non-Goals:**
    *   Implementing a full OAuth2 server for local-only use.
    *   Protecting against a fully compromised host (if the OS is compromised, the gateway is compromised).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local Developer / Agent Orchestrator
* **Primary Goal:** Use a browser-based agent dashboard to interact with local MCP Any tools without exposing the gateway to general web-based CSRF attacks.
* **The Happy Path (Tasks):**
    1.  User starts `mcpany server`.
    2.  Server generates a unique `Session-Secret` stored in a local-only file (e.g., `~/.mcpany/auth_token`).
    3.  Authorized Dashboard (running on `localhost:3000` or a trusted domain) reads the secret or receives it via a secure local handshake.
    4.  Dashboard includes an `X-MCP-Auth-Token` or a signed Nonce in every request.
    5.  `mcpany` gateway validates the token/signature and the `Origin` header before processing the tool call.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Browser as Malicious Website
        participant Client as Trusted Dashboard
        participant Gateway as MCP Any Gateway

        Note over Browser, Gateway: CSRF Attempt
        Browser->>Gateway: POST /tool/execute (No Token)
        Gateway-->>Browser: 403 Forbidden (Missing/Invalid Token)

        Note over Client, Gateway: Authorized Flow
        Client->>Gateway: GET /auth/handshake (Local-only)
        Gateway-->>Client: Returns Nonce
        Client->>Gateway: POST /tool/execute (X-MCP-Nonce: Signed-Nonce)
        Gateway->>Gateway: Validate Signature & Origin
        Gateway-->>Client: 200 OK (Tool Output)
    ```
* **APIs / Interfaces:**
    *   `GET /v1/auth/token`: Returns a short-lived token for authorized local clients. Restricted to `127.0.0.1`.
    *   Middleware: `CSRFGuard` - Injected into the HTTP pipeline to intercept all mutating requests.
* **Data Storage/State:**
    *   `Session-Secret` is ephemeral and regenerated on server restart.
    *   `Authorized-Origins` allowlist maintained in `config.yaml`.

## 5. Alternatives Considered
* **Static API Keys:** Rejected because they are often committed to git or leaked in logs. Ephemeral session tokens are more secure for local-use cases.
* **OIDC/OAuth:** Too much overhead for a simple local gateway, though potentially useful for the Federated Mesh future.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The gateway must default to `Deny All` for any request missing the cryptographic proof of local presence.
* **Observability:** Log all blocked CSRF attempts with the `Referer` and `Origin` headers for debugging and threat intelligence.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
